package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/neto-app/neto/api/internal/domain/entity"
)

// AIConversationRepository define las operaciones de persistencia para conversaciones de IA.
type AIConversationRepository interface {
	Create(ctx context.Context, conv entity.AIConversation) error
	GetByID(ctx context.Context, id uuid.UUID) (entity.AIConversation, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.AIConversation, error)
	UpdateTitle(ctx context.Context, id uuid.UUID, title string) error
}

// AIMessageRepository define las operaciones de persistencia para mensajes de conversaciones de IA.
type AIMessageRepository interface {
	Create(ctx context.Context, msg entity.AIMessage) error
	GetByConversationID(ctx context.Context, conversationID uuid.UUID) ([]entity.AIMessage, error)
}
