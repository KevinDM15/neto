// Package ai defines the port (abstraction) for any language model provider.
// It has ZERO external dependencies — only stdlib.
// Implementations live in internal/infrastructure/{provider}/.
package ai

import (
	"context"
	"encoding/json"
	"errors"
)

// ErrMaxIterations is returned when the tool-use loop exceeds the maximum number of iterations.
var ErrMaxIterations = errors.New("ai: max tool-use iterations reached")

// Role represents the role of a message in the conversation.
type Role string

const (
	// RoleUser is the role of the user message.
	RoleUser Role = "user"
	// RoleAssistant is the role of the assistant response.
	RoleAssistant Role = "assistant"
	// RoleTool is the role used for tool result messages.
	RoleTool Role = "tool"
)

// ContentBlock represents a content block within a message.
// It can be text, tool_use, or tool_result.
type ContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
}

// Message represents a message in the conversation with an LLM provider.
type Message struct {
	Role    Role            `json:"role"`
	Content json.RawMessage `json:"content"`
}

// NewTextMessage creates a simple text Message.
func NewTextMessage(role Role, text string) (Message, error) {
	blocks := []ContentBlock{{Type: "text", Text: text}}
	raw, err := json.Marshal(blocks)
	if err != nil {
		return Message{}, err
	}
	return Message{Role: role, Content: raw}, nil
}

// Tool defines a tool available to the language model.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// LoopResult is the result of the tool-use loop.
// Exactly one of Reply or PendingConfirmation will be set.
type LoopResult struct {
	// Reply contains the final text response when the loop ends normally.
	Reply string
	// PendingConfirmation contains the tool_use block that requires user approval.
	PendingConfirmation *ContentBlock
	// Messages is the full message history after the loop.
	Messages []Message
}

// ToolExecutorFunc is the function that executes a tool_use block.
// It returns (result json, isPending, error).
// isPending=true means the tool requires user confirmation before executing.
type ToolExecutorFunc func(ctx context.Context, block *ContentBlock) (json.RawMessage, bool, error)

// LLMClient is the port for any language model provider.
// Implementations live in internal/infrastructure/{provider}/.
type LLMClient interface {
	// RunToolLoop sends messages to the model and executes tool calls until
	// the model stops, a confirmation is required, or the iteration limit is reached.
	RunToolLoop(ctx context.Context, messages []Message, tools []Tool, exec ToolExecutorFunc) (*LoopResult, error)

	// WithSystemPrompt returns a new client configured with the given system prompt.
	WithSystemPrompt(prompt string) LLMClient
}
