// Package middleware contiene los middlewares HTTP de la API de Neto.
package middleware

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// contextKey es el tipo privado para las keys del context — evita colisiones.
type contextKey string

const (
	// contextKeyUserID es la key para el UUID del usuario autenticado.
	contextKeyUserID contextKey = "userID"
)

// UserIDFromContext extrae el UUID del usuario autenticado del context.
// Retorna error si el context no contiene un userID válido.
func UserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	val := ctx.Value(contextKeyUserID)
	if val == nil {
		return uuid.Nil, errors.New("user id not found in context")
	}
	id, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user id in context has invalid type")
	}
	return id, nil
}

// withUserID inyecta el userID en el context.
func withUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, contextKeyUserID, id)
}

// UserIDCtxKey retorna la key del context usada para el userID.
// Se expone para uso en tests — no usar en producción fuera del middleware.
func UserIDCtxKey() any {
	return contextKeyUserID
}
