package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neto-app/neto/api/internal/domain/entity"
)

// AIConversationRepository implementa domain/repository.AIConversationRepository usando pgx.
type AIConversationRepository struct {
	pool *pgxpool.Pool
}

// newAIConversationRepository crea un AIConversationRepository con el pool dado.
func newAIConversationRepository(pool *pgxpool.Pool) *AIConversationRepository {
	return &AIConversationRepository{pool: pool}
}

// Create inserta una nueva conversación en la DB.
func (r *AIConversationRepository) Create(ctx context.Context, conv entity.AIConversation) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO ai_conversations (id, user_id, title, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`,
		conv.ID,
		conv.UserID,
		conv.Title,
		conv.CreatedAt,
		conv.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository: create ai_conversation: %w", err)
	}
	return nil
}

// GetByID busca una conversación por ID.
func (r *AIConversationRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.AIConversation, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, title, created_at, updated_at
		FROM ai_conversations
		WHERE id = $1
	`, id)

	var conv entity.AIConversation
	var title *string
	if err := row.Scan(&conv.ID, &conv.UserID, &title, &conv.CreatedAt, &conv.UpdatedAt); err != nil {
		return entity.AIConversation{}, fmt.Errorf("repository: get ai_conversation: %w", err)
	}
	if title != nil {
		conv.Title = *title
	}
	return conv, nil
}

// GetByUserID retorna todas las conversaciones del usuario.
func (r *AIConversationRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.AIConversation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, title, created_at, updated_at
		FROM ai_conversations
		WHERE user_id = $1
		ORDER BY updated_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: get ai_conversations by user: %w", err)
	}
	defer rows.Close()

	var convs []entity.AIConversation
	for rows.Next() {
		var conv entity.AIConversation
		var title *string
		if err := rows.Scan(&conv.ID, &conv.UserID, &title, &conv.CreatedAt, &conv.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repository: scan ai_conversation: %w", err)
		}
		if title != nil {
			conv.Title = *title
		}
		convs = append(convs, conv)
	}
	return convs, rows.Err()
}

// UpdateTitle actualiza el título de una conversación.
func (r *AIConversationRepository) UpdateTitle(ctx context.Context, id uuid.UUID, title string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE ai_conversations
		SET title = $1, updated_at = $2
		WHERE id = $3
	`, title, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("repository: update ai_conversation title: %w", err)
	}
	return nil
}

// AIMessageRepository implementa domain/repository.AIMessageRepository usando pgx.
type AIMessageRepository struct {
	pool *pgxpool.Pool
}

// newAIMessageRepository crea un AIMessageRepository con el pool dado.
func newAIMessageRepository(pool *pgxpool.Pool) *AIMessageRepository {
	return &AIMessageRepository{pool: pool}
}

// Create inserta un nuevo mensaje en la DB.
func (r *AIMessageRepository) Create(ctx context.Context, msg entity.AIMessage) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO ai_messages (id, conversation_id, role, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`,
		msg.ID,
		msg.ConversationID,
		string(msg.Role),
		[]byte(msg.Content),
		msg.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository: create ai_message: %w", err)
	}
	return nil
}

// GetByConversationID retorna todos los mensajes de una conversación ordenados por fecha.
func (r *AIMessageRepository) GetByConversationID(ctx context.Context, conversationID uuid.UUID) ([]entity.AIMessage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, conversation_id, role, content, created_at
		FROM ai_messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
	`, conversationID)
	if err != nil {
		return nil, fmt.Errorf("repository: get ai_messages by conversation: %w", err)
	}
	defer rows.Close()

	var msgs []entity.AIMessage
	for rows.Next() {
		var msg entity.AIMessage
		var role string
		var content []byte
		if err := rows.Scan(&msg.ID, &msg.ConversationID, &role, &content, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("repository: scan ai_message: %w", err)
		}
		msg.Role = entity.AIMessageRole(role)
		msg.Content = json.RawMessage(content)
		msgs = append(msgs, msg)
	}
	return msgs, rows.Err()
}
