// Package openrouter provee un cliente HTTP para la API de OpenRouter.
// OpenRouter expone una API compatible con OpenAI en https://openrouter.ai/api/v1.
// Implementa ai.LLMClient como adaptador de infraestructura.
package openrouter

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
	// DefaultModel es el modelo por defecto — Gemini 2.0 Flash (free tier).
	DefaultModel = "google/gemini-2.0-flash-exp:free"

	// apiURL es el endpoint de chat completions de OpenRouter.
	apiURL = "https://openrouter.ai/api/v1/chat/completions"

	// maxIterations limita el loop de tool use para evitar ciclos infinitos.
	maxIterations = 5
)

// orFunction representa la definición de una función en el formato OpenRouter/OpenAI.
type orFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// orTool representa un tool en el formato OpenRouter/OpenAI.
type orTool struct {
	Type     string     `json:"type"`
	Function orFunction `json:"function"`
}

// orFunctionCall representa la llamada a una función dentro de un tool_call.
type orFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// orToolCall representa una llamada a un tool en la respuesta del asistente.
type orToolCall struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Function orFunctionCall `json:"function"`
}

// orMessage representa un mensaje en el formato OpenRouter/OpenAI.
type orMessage struct {
	Role       string       `json:"role"`
	Content    any          `json:"content"`
	ToolCalls  []orToolCall `json:"tool_calls,omitempty"`
	ToolCallID string       `json:"tool_call_id,omitempty"`
}

// orRequest es el body enviado a la API de OpenRouter.
type orRequest struct {
	Model    string      `json:"model"`
	Messages []orMessage `json:"messages"`
	Tools    []orTool    `json:"tools,omitempty"`
}

// orChoice representa una opción en la respuesta de OpenRouter.
type orChoice struct {
	FinishReason string    `json:"finish_reason"`
	Message      orMessage `json:"message"`
}

// orResponse es la respuesta de la API de OpenRouter.
type orResponse struct {
	ID      string     `json:"id"`
	Choices []orChoice `json:"choices"`
}

// orErrorResponse es la respuesta de error de la API.
type orErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

// Client es el cliente HTTP para la API de OpenRouter.
// Implementa ai.LLMClient.
type Client struct {
	apiKey       string
	model        string
	httpClient   *http.Client
	systemPrompt string
}

// NewClient crea un nuevo Client con el API key y modelo dados.
func NewClient(apiKey, model string) *Client {
	if model == "" {
		model = DefaultModel
	}
	return &Client{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// WithSystemPrompt configura el system prompt del cliente y retorna un nuevo ai.LLMClient.
func (c *Client) WithSystemPrompt(prompt string) ai.LLMClient {
	clone := *c
	clone.systemPrompt = prompt
	return &clone
}

// RunToolLoop ejecuta el loop de tool use con máximo maxIterations iteraciones.
// Convierte los mensajes ai.Message al formato OpenRouter y ejecuta los tool calls.
func (c *Client) RunToolLoop(
	ctx context.Context,
	messages []ai.Message,
	tools []ai.Tool,
	executor ai.ToolExecutorFunc,
) (*ai.LoopResult, error) {
	orMsgs, err := convertMessages(messages, c.systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("openrouter: convert messages: %w", err)
	}

	orTools := convertTools(tools)

	// Guardamos los ai.Message originales para retornar en LoopResult
	aiMsgs := make([]ai.Message, len(messages))
	copy(aiMsgs, messages)

	for i := range maxIterations {
		resp, err := c.sendRequest(ctx, orMsgs, orTools)
		if err != nil {
			return nil, fmt.Errorf("openrouter: iteration %d: %w", i, err)
		}

		if len(resp.Choices) == 0 {
			return nil, fmt.Errorf("openrouter: empty choices in response")
		}

		choice := resp.Choices[0]
		assistantMsg := choice.Message

		// Agregar el mensaje del asistente al historial
		orMsgs = append(orMsgs, assistantMsg)

		switch choice.FinishReason {
		case "stop":
			content, _ := assistantMsg.Content.(string)
			return &ai.LoopResult{Reply: content, Messages: aiMsgs}, nil

		case "tool_calls":
			var pendingBlock *ai.ContentBlock

			for _, tc := range assistantMsg.ToolCalls {
				// Parsear los argumentos del tool call
				var input json.RawMessage
				if tc.Function.Arguments == "" || tc.Function.Arguments == "null" {
					input = json.RawMessage("{}")
				} else {
					input = json.RawMessage(tc.Function.Arguments)
				}

				block := &ai.ContentBlock{
					Type:  "tool_use",
					ID:    tc.ID,
					Name:  tc.Function.Name,
					Input: input,
				}

				result, pending, err := executor(ctx, block)
				if err != nil {
					return nil, fmt.Errorf("openrouter: execute tool %q: %w", block.Name, err)
				}

				if pending {
					pendingBlock = block
					break
				}

				// Agregar resultado como mensaje de tool al historial
				resultStr := string(result)
				orMsgs = append(orMsgs, orMessage{
					Role:       "tool",
					Content:    resultStr,
					ToolCallID: tc.ID,
				})
			}

			if pendingBlock != nil {
				return &ai.LoopResult{
					PendingConfirmation: pendingBlock,
					Messages:            aiMsgs,
				}, nil
			}

		default:
			// Razón de parada inesperada — retornar lo que tengamos
			content, _ := assistantMsg.Content.(string)
			return &ai.LoopResult{Reply: content, Messages: aiMsgs}, nil
		}
	}

	return nil, ai.ErrMaxIterations
}

// sendRequest envía un request a la API de OpenRouter y retorna la respuesta.
func (c *Client) sendRequest(ctx context.Context, messages []orMessage, tools []orTool) (*orResponse, error) {
	reqBody := orRequest{
		Model:    c.model,
		Messages: messages,
		Tools:    tools,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/KevinDM15/neto")
	req.Header.Set("X-Title", "Neto")

	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode != http.StatusOK {
		var errResp orErrorResponse
		if decodeErr := json.NewDecoder(httpResp.Body).Decode(&errResp); decodeErr == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("api error %d: %s", httpResp.StatusCode, errResp.Error.Message)
		}
		return nil, fmt.Errorf("api error: status %d", httpResp.StatusCode)
	}

	var resp orResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// convertMessages convierte []ai.Message al formato de mensajes de OpenRouter.
// Si systemPrompt no está vacío, se antepone como mensaje de sistema.
func convertMessages(messages []ai.Message, systemPrompt string) ([]orMessage, error) {
	var result []orMessage

	if systemPrompt != "" {
		result = append(result, orMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	for _, msg := range messages {
		// Intentar deserializar como []ContentBlock primero
		var blocks []ai.ContentBlock
		if err := json.Unmarshal(msg.Content, &blocks); err != nil {
			// Si falla, intentar como string
			var text string
			if err2 := json.Unmarshal(msg.Content, &text); err2 != nil {
				return nil, fmt.Errorf("convert message role=%s: %w", msg.Role, err)
			}
			result = append(result, orMessage{Role: string(msg.Role), Content: text})
			continue
		}

		converted, err := convertBlocksToOR(msg.Role, blocks)
		if err != nil {
			return nil, err
		}
		result = append(result, converted...)
	}

	return result, nil
}

// convertBlocksToOR convierte []ai.ContentBlock a mensajes de OpenRouter según el rol.
func convertBlocksToOR(role ai.Role, blocks []ai.ContentBlock) ([]orMessage, error) {
	switch role {
	case ai.RoleUser:
		// Extraer texto de los bloques de usuario
		text := ""
		for _, b := range blocks {
			if b.Type == "text" {
				text = b.Text
				break
			}
		}
		return []orMessage{{Role: "user", Content: text}}, nil

	case ai.RoleAssistant:
		msg := orMessage{Role: "assistant", Content: nil}
		for _, b := range blocks {
			switch b.Type {
			case "text":
				msg.Content = b.Text
			case "tool_use":
				args := "{}"
				if len(b.Input) > 0 {
					args = string(b.Input)
				}
				msg.ToolCalls = append(msg.ToolCalls, orToolCall{
					ID:   b.ID,
					Type: "function",
					Function: orFunctionCall{
						Name:      b.Name,
						Arguments: args,
					},
				})
			}
		}
		return []orMessage{msg}, nil

	case ai.RoleTool:
		// Cada tool_result se convierte en un mensaje "tool" separado
		var msgs []orMessage
		for _, b := range blocks {
			if b.Type != "tool_result" {
				continue
			}
			content := ""
			if len(b.Content) > 0 {
				// Si es un string JSON, desempaquetar
				var s string
				if err := json.Unmarshal(b.Content, &s); err == nil {
					content = s
				} else {
					content = string(b.Content)
				}
			}
			msgs = append(msgs, orMessage{
				Role:       "tool",
				Content:    content,
				ToolCallID: b.ToolUseID,
			})
		}
		return msgs, nil

	default:
		return nil, fmt.Errorf("unsupported role: %s", role)
	}
}

// convertTools convierte []ai.Tool al formato de tools de OpenRouter.
func convertTools(tools []ai.Tool) []orTool {
	result := make([]orTool, 0, len(tools))
	for _, t := range tools {
		result = append(result, orTool{
			Type: "function",
			Function: orFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.InputSchema,
			},
		})
	}
	return result
}
