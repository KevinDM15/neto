// Package anthropic provee un cliente HTTP para la API de Mensajes de Anthropic.
// No usa streaming — respuestas JSON completas solamente.
// Implementa ai.LLMClient como adaptador de infraestructura.
package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/neto-app/neto/api/internal/ai"
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

// anthropicRequest es el body enviado a la API de Anthropic.
type anthropicRequest struct {
	Model     string       `json:"model"`
	MaxTokens int          `json:"max_tokens"`
	System    string       `json:"system,omitempty"`
	Messages  []ai.Message `json:"messages"`
	Tools     []ai.Tool    `json:"tools,omitempty"`
}

// anthropicResponse es la respuesta de la API de Anthropic.
type anthropicResponse struct {
	ID           string            `json:"id"`
	Type         string            `json:"type"`
	Role         string            `json:"role"`
	Content      []ai.ContentBlock `json:"content"`
	StopReason   string            `json:"stop_reason"`
	StopSequence string            `json:"stop_sequence"`
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
// Implementa ai.LLMClient.
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

// WithSystemPrompt configura el system prompt del cliente y retorna un nuevo ai.LLMClient.
func (c *Client) WithSystemPrompt(prompt string) ai.LLMClient {
	clone := *c
	clone.systemPrompt = prompt
	return &clone
}

// RunToolLoop ejecuta el loop de tool use con máximo maxIterations iteraciones.
// El executor recibe el nombre del tool y el input crudo y retorna el resultado.
// Si un tool requiere confirmación, retorna LoopResult{PendingConfirmation: &block}.
func (c *Client) RunToolLoop(
	ctx context.Context,
	messages []ai.Message,
	tools []ai.Tool,
	executor ai.ToolExecutorFunc,
) (*ai.LoopResult, error) {
	msgs := make([]ai.Message, len(messages))
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
		assistantMsg := ai.Message{Role: ai.RoleAssistant, Content: assistantContent}
		msgs = append(msgs, assistantMsg)

		// Si Claude terminó sin tool_use, retornar el texto final
		if resp.StopReason == "end_turn" {
			text := extractText(resp.Content)
			return &ai.LoopResult{Reply: text, Messages: msgs}, nil
		}

		// Procesar tool_use blocks
		if resp.StopReason != "tool_use" {
			// Razón de parada inesperada — retornar lo que tengamos
			text := extractText(resp.Content)
			return &ai.LoopResult{Reply: text, Messages: msgs}, nil
		}

		// Ejecutar cada tool_use y recolectar resultados
		var toolResults []ai.ContentBlock
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
				return &ai.LoopResult{
					PendingConfirmation: block,
					Messages:            msgs,
				}, nil
			}

			toolResults = append(toolResults, ai.ContentBlock{
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
			msgs = append(msgs, ai.Message{Role: ai.RoleUser, Content: resultContent})
		}
	}

	return nil, ai.ErrMaxIterations
}

// sendRequest envía un request a la API de Anthropic y retorna la respuesta.
func (c *Client) sendRequest(ctx context.Context, messages []ai.Message, tools []ai.Tool) (*anthropicResponse, error) {
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
func extractText(blocks []ai.ContentBlock) string {
	for _, b := range blocks {
		if b.Type == "text" {
			return b.Text
		}
	}
	return ""
}
