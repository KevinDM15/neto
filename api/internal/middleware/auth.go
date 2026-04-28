package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Authenticator retorna un middleware que valida el JWT de Supabase.
// Extrae el sub (user UUID) y lo inyecta en el context.
// Retorna 401 si el token está ausente o es inválido.
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

			ctx := withUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
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

// parseJWT valida y parsea el JWT con el secret de Supabase.
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
