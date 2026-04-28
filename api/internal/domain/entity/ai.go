package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AIConversation representa una sesión de conversación entre un usuario y el agente Neto.
type AIConversation struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AIMessageRole define los posibles roles de un mensaje en la conversación.
type AIMessageRole string

const (
	// AIRoleUser es el rol del mensaje del usuario.
	AIRoleUser AIMessageRole = "user"
	// AIRoleAssistant es el rol de la respuesta del asistente.
	AIRoleAssistant AIMessageRole = "assistant"
	// AIRoleToolUse es el rol de un mensaje que contiene un tool_use block.
	AIRoleToolUse AIMessageRole = "tool_use"
	// AIRoleToolResult es el rol de un mensaje que contiene el resultado de un tool.
	AIRoleToolResult AIMessageRole = "tool_result"
)

// AIMessage representa un mensaje individual dentro de una conversación.
type AIMessage struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	Role           AIMessageRole
	Content        json.RawMessage
	CreatedAt      time.Time
}
