// Package anthropic provee un cliente HTTP para la API de Mensajes de Anthropic.
// No usa streaming — respuestas JSON completas solamente.
package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	// DefaultModel es el modelo por defecto — rápido y económico.
	DefaultModel = "claude-3-5-haiku-20241022"

	// anthropicAPIURL es el endpoint de mensajes de Anthropic.
	anthropicAPIURL = "https://api.anthropic.com/v1/messages"

	// anthropicVersion es la versión de la API requerida en el header.
	anthropicVersion = "2023-06-01"

	// maxIterations limita el loop de tool use para evitar ciclos infinitos.
	maxIterations = 5
)

// ErrMaxIterations se retorna cuando se supera el límite de iteraciones del loop.
var ErrMaxIterations = errors.New("anthropic: max tool-use iterations reached")

// Role representa el rol de un mensaje en la conversación.
type Role string

const (
	// RoleUser es el rol del mensaje del usuario.
	RoleUser Role = "user"
	// RoleAssistant es el rol de la respuesta del asistente.
	RoleAssistant Role = "assistant"
)

// ContentBlock representa un bloque de contenido dentro de un mensaje.
// Puede ser text, tool_use o tool_result.
type ContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
}

// Message representa un mensaje en la conversación con Anthropic.
type Message struct {
	Role    Role            `json:"role"`
	Content json.RawMessage `json:"content"`
}

// NewTextMessage crea un Message de texto simple.
func NewTextMessage(role Role, text string) (Message, error) {
	blocks := []ContentBlock{{Type: "text", Text: text}}
	raw, err := json.Marshal(blocks)
	if err != nil {
		return Message{}, fmt.Errorf("anthropic: marshal text message: %w", err)
	}
	return Message{Role: role, Content: raw}, nil
}

// Tool define una herramienta disponible para Claude.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// LoopResult es el resultado del loop de tool use.
// Exactamente uno de Reply o PendingConfirmation estará definido.
type LoopResult struct {
	// Reply contiene la respuesta final en texto cuando el loop termina normalmente.
	Reply string
	// PendingConfirmation contiene el bloque de tool_use que requiere aprobación del usuario.
	PendingConfirmation *ContentBlock
	// Messages es el historial completo de mensajes tras el loop.
	Messages []Message
}

// anthropicRequest es el body enviado a la API de Anthropic.
type anthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
	Tools     []Tool    `json:"tools,omitempty"`
}

// anthropicResponse es la respuesta de la API de Anthropic.
type anthropicResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	StopReason   string         `json:"stop_reason"`
	StopSequence string         `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// anthropicErrorResponse es la respuesta de error de la API.
type anthropicErrorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// Client es el cliente HTTP para la API de Anthropic.
type Client struct {
	apiKey       string
	httpClient   *http.Client
	model        string
	systemPrompt string
}

// NewClient crea un nuevo Client con el API key dado.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		model: DefaultModel,
	}
}

// WithSystemPrompt configura el system prompt del cliente.
func (c *Client) WithSystemPrompt(prompt string) *Client {
	c.systemPrompt = prompt
	return c
}

// RunToolLoop ejecuta el loop de tool use con máximo maxIterations iteraciones.
// El executor recibe el nombre del tool y el input crudoy retorna el resultado.
// Si un tool requiere confirmación, retorna LoopResult{PendingConfirmation: &block}.
func (c *Client) RunToolLoop(
	ctx context.Context,
	messages []Message,
	tools []Tool,
	executor ToolExecutorFunc,
) (*LoopResult, error) {
	msgs := make([]Message, len(messages))
	copy(msgs, messages)

	for i := range maxIterations {
		resp, err := c.sendRequest(ctx, msgs, tools)
		if err != nil {
			return nil, fmt.Errorf("anthropic: iteration %d: %w", i, err)
		}

		// Construir el mensaje del asistente con su contenido completo
		assistantContent, err := json.Marshal(resp.Content)
		if err != nil {
			return nil, fmt.Errorf("anthropic: marshal assistant content: %w", err)
		}
		assistantMsg := Message{Role: RoleAssistant, Content: assistantContent}
		msgs = append(msgs, assistantMsg)

		// Si Claude terminó sin tool_use, retornar el texto final
		if resp.StopReason == "end_turn" {
			text := extractText(resp.Content)
			return &LoopResult{Reply: text, Messages: msgs}, nil
		}

		// Procesar tool_use blocks
		if resp.StopReason != "tool_use" {
			// Razón de parada inesperada — retornar lo que tengamos
			text := extractText(resp.Content)
			return &LoopResult{Reply: text, Messages: msgs}, nil
		}

		// Ejecutar cada tool_use y recolectar resultados
		var toolResults []ContentBlock
		for idx := range resp.Content {
			block := &resp.Content[idx]
			if block.Type != "tool_use" {
				continue
			}

			result, pending, err := executor(ctx, block)
			if err != nil {
				return nil, fmt.Errorf("anthropic: execute tool %q: %w", block.Name, err)
			}

			// Si el tool requiere confirmación, pausar el loop
			if pending {
				return &LoopResult{
					PendingConfirmation: block,
					Messages:            msgs,
				}, nil
			}

			toolResults = append(toolResults, ContentBlock{
				Type:      "tool_result",
				ToolUseID: block.ID,
				Content:   result,
			})
		}

		// Agregar los resultados al historial como mensaje de usuario
		if len(toolResults) > 0 {
			resultContent, err := json.Marshal(toolResults)
			if err != nil {
				return nil, fmt.Errorf("anthropic: marshal tool results: %w", err)
			}
			msgs = append(msgs, Message{Role: RoleUser, Content: resultContent})
		}
	}

	return nil, ErrMaxIterations
}

// sendRequest envía un request a la API de Anthropic y retorna la respuesta.
func (c *Client) sendRequest(ctx context.Context, messages []Message, tools []Tool) (*anthropicResponse, error) {
	reqBody := anthropicRequest{
		Model:     c.model,
		MaxTokens: 4096,
		System:    c.systemPrompt,
		Messages:  messages,
		Tools:     tools,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPIURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)

	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode != http.StatusOK {
		var errResp anthropicErrorResponse
		if decodeErr := json.NewDecoder(httpResp.Body).Decode(&errResp); decodeErr == nil {
			return nil, fmt.Errorf("api error %d: %s", httpResp.StatusCode, errResp.Error.Message)
		}
		return nil, fmt.Errorf("api error: status %d", httpResp.StatusCode)
	}

	var resp anthropicResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// extractText extrae el primer bloque de texto de un slice de ContentBlock.
func extractText(blocks []ContentBlock) string {
	for _, b := range blocks {
		if b.Type == "text" {
			return b.Text
		}
	}
	return ""
}

// ToolExecutorFunc es la función que ejecuta un tool_use block.
// Retorna (result json, isPending, error).
// isPending=true significa que el tool requiere confirmación del usuario.
type ToolExecutorFunc func(ctx context.Context, block *ContentBlock) (json.RawMessage, bool, error)
