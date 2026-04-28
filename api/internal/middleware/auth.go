package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWKSAuthenticator retorna un middleware que valida JWTs de Supabase
// usando el JWKS endpoint. Soporta cualquier algoritmo publicado en el JWKS
// (HS256 legacy, ES256 nuevo), sin hardcodear un secret.
func JWKSAuthenticator(jwksURL string) (func(http.Handler) http.Handler, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	k, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("fetch JWKS from %s: %w", jwksURL, err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr, err := extractBearerToken(r)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing or malformed authorization header"})
				return
			}

			token, err := jwt.Parse(tokenStr, k.Keyfunc)
			if err != nil || !token.Valid {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
				return
			}

			sub, err := token.Claims.GetSubject()
			if err != nil || sub == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "token missing subject"})
				return
			}

			userID, err := uuid.Parse(sub)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid user id in token"})
				return
			}

			next.ServeHTTP(w, r.WithContext(withUserID(r.Context(), userID)))
		})
	}, nil
}

// Authenticator es el middleware legacy (HS256 con secret compartido).
// Mantenido para tests unitarios que no tienen acceso a un JWKS real.
func Authenticator(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := extractBearerToken(r)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing or malformed authorization header"})
				return
			}

			claims, err := parseJWT(token, jwtSecret)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
				return
			}

			sub, err := claims.GetSubject()
			if err != nil || sub == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "token missing subject"})
				return
			}

			userID, err := uuid.Parse(sub)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid user id in token"})
				return
			}

			next.ServeHTTP(w, r.WithContext(withUserID(r.Context(), userID)))
		})
	}
}

// extractBearerToken extrae el token del header Authorization: Bearer <token>.
func extractBearerToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", jwt.ErrTokenMalformed
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", jwt.ErrTokenMalformed
	}

	return parts[1], nil
}

// parseJWT valida y parsea un JWT con HMAC (HS256). Solo para tests.
func parseJWT(tokenStr, secret string) (jwt.Claims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrTokenSignatureInvalid
	}
	return token.Claims, nil
}

// writeJSON escribe una respuesta JSON con el status code dado.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
