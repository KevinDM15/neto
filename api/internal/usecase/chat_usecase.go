package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/neto-app/neto/api/internal/ai"
	"github.com/neto-app/neto/api/internal/domain/entity"
	domainrepo "github.com/neto-app/neto/api/internal/domain/repository"
)

// ChatRequest agrupa los datos de una solicitud al agente.
type ChatRequest struct {
	ConversationID *uuid.UUID
	Message        string
	Confirm        bool
	PendingTool    *ai.ContentBlock
}

// PendingConfirmationResponse contiene los datos de un tool que requiere confirmación.
type PendingConfirmationResponse struct {
	Tool    string          `json:"tool"`
	Preview json.RawMessage `json:"preview"`
	Block   *ai.ContentBlock `json:"block"`
}

// ChatResponse es la respuesta del agente al usuario.
type ChatResponse struct {
	ConversationID      uuid.UUID                    `json:"conversation_id"`
	Reply               string                       `json:"reply,omitempty"`
	PendingConfirmation *PendingConfirmationResponse `json:"pending_confirmation,omitempty"`
}

// ChatUseCase orquesta el loop de conversación con el agente Neto.
type ChatUseCase struct {
	llm         ai.LLMClient
	executor    ToolExecutorBuilder
	tools       []ai.Tool
	convRepo    domainrepo.AIConversationRepository
	messageRepo domainrepo.AIMessageRepository
}

// NewChatUseCase crea un nuevo ChatUseCase.
func NewChatUseCase(
	llm ai.LLMClient,
	executor ToolExecutorBuilder,
	tools []ai.Tool,
	convRepo domainrepo.AIConversationRepository,
	messageRepo domainrepo.AIMessageRepository,
) *ChatUseCase {
	return &ChatUseCase{
		llm:         llm,
		executor:    executor,
		tools:       tools,
		convRepo:    convRepo,
		messageRepo: messageRepo,
	}
}

// Chat procesa un mensaje del usuario y retorna la respuesta del agente.
func (uc *ChatUseCase) Chat(ctx context.Context, userID uuid.UUID, req ChatRequest) (ChatResponse, error) {
	// Obtener o crear conversación
	convID, err := uc.resolveConversation(ctx, userID, req.ConversationID)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("chat: resolve conversation: %w", err)
	}

	// Cargar historial de mensajes
	history, err := uc.loadHistory(ctx, convID)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("chat: load history: %w", err)
	}

	// Ejecutar el loop de tool use
	executorFn := uc.executor.BuildExecutorFunc(userID, req.Confirm)

	// Construir el mensaje del usuario
	var msgs []ai.Message

	// Si hay un pending tool siendo confirmado, ejecutarlo y reinyectar el estado previo
	if req.Confirm && req.PendingTool != nil {
		msgs = history
		
		// Ejecutar el tool manualmente ya que el usuario lo confirmó
		resultJson, _, err := executorFn(ctx, req.PendingTool)
		if err != nil {
			return ChatResponse{}, fmt.Errorf("chat: execute confirmed tool %q: %w", req.PendingTool.Name, err)
		}
		
		// Armar el mensaje de tool result
		toolResultBlock := ai.ContentBlock{
			Type:      "tool_result",
			ToolUseID: req.PendingTool.ID,
			Content:   resultJson,
		}
		
		rawBlock, err := json.Marshal([]ai.ContentBlock{toolResultBlock})
		if err != nil {
			return ChatResponse{}, fmt.Errorf("chat: marshal tool result: %w", err)
		}

		toolResultMsg := ai.Message{
			Role:    ai.RoleTool,
			Content: rawBlock,
		}
		
		msgs = append(msgs, toolResultMsg)
		
		// Persistir el resultado para que el historial lo refleje y el LLM lo vea
		if err := uc.saveMessage(ctx, convID, entity.AIRoleToolUse, rawBlock); err != nil {
			return ChatResponse{}, fmt.Errorf("chat: save tool result: %w", err)
		}
	} else {
		// Mensaje nuevo del usuario
		userMsg, err := ai.NewTextMessage(ai.RoleUser, req.Message)
		if err != nil {
			return ChatResponse{}, fmt.Errorf("chat: build user message: %w", err)
		}
		msgs = append(history, userMsg)

		// Persistir el mensaje del usuario
		if err := uc.saveMessage(ctx, convID, entity.AIRoleUser, userMsg.Content); err != nil {
			return ChatResponse{}, fmt.Errorf("chat: save user message: %w", err)
		}
	}

	result, err := uc.llm.RunToolLoop(ctx, msgs, uc.tools, executorFn)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("chat: run tool loop: %w", err)
	}

	// Persistir los mensajes nuevos del loop
	if err := uc.persistLoopMessages(ctx, convID, msgs, result.Messages); err != nil {
		return ChatResponse{}, fmt.Errorf("chat: persist loop messages: %w", err)
	}

	// Construir la respuesta
	resp := ChatResponse{ConversationID: convID}

	if result.PendingConfirmation != nil {
		resp.PendingConfirmation = &PendingConfirmationResponse{
			Tool:    result.PendingConfirmation.Name,
			Preview: result.PendingConfirmation.Input,
			Block:   result.PendingConfirmation,
		}
		return resp, nil
	}

	resp.Reply = result.Reply
	return resp, nil
}

// resolveConversation retorna el ID de conversación existente o crea una nueva.
func (uc *ChatUseCase) resolveConversation(ctx context.Context, userID uuid.UUID, convID *uuid.UUID) (uuid.UUID, error) {
	if convID != nil {
		// Verificar que la conversación exista
		conv, err := uc.convRepo.GetByID(ctx, *convID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("get conversation: %w", err)
		}
		if conv.UserID != userID {
			return uuid.Nil, fmt.Errorf("conversation does not belong to user")
		}
		return *convID, nil
	}

	// Crear nueva conversación
	now := time.Now().UTC()
	conv := entity.AIConversation{
		ID:        uuid.New(),
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := uc.convRepo.Create(ctx, conv); err != nil {
		return uuid.Nil, fmt.Errorf("create conversation: %w", err)
	}
	return conv.ID, nil
}

// loadHistory carga el historial de mensajes de una conversación como []ai.Message.
func (uc *ChatUseCase) loadHistory(ctx context.Context, convID uuid.UUID) ([]ai.Message, error) {
	msgs, err := uc.messageRepo.GetByConversationID(ctx, convID)
	if err != nil {
		return nil, fmt.Errorf("load history: %w", err)
	}

	var history []ai.Message
	for _, m := range msgs {
		role := ai.RoleUser
		if m.Role == entity.AIRoleAssistant || m.Role == entity.AIRoleToolUse {
			role = ai.RoleAssistant
		}
		history = append(history, ai.Message{Role: role, Content: m.Content})
	}
	return history, nil
}

// saveMessage persiste un mensaje en la DB.
func (uc *ChatUseCase) saveMessage(ctx context.Context, convID uuid.UUID, role entity.AIMessageRole, content json.RawMessage) error {
	msg := entity.AIMessage{
		ID:             uuid.New(),
		ConversationID: convID,
		Role:           role,
		Content:        content,
		CreatedAt:      time.Now().UTC(),
	}
	return uc.messageRepo.Create(ctx, msg)
}

// persistLoopMessages guarda los mensajes nuevos generados durante el loop.
func (uc *ChatUseCase) persistLoopMessages(ctx context.Context, convID uuid.UUID, before, after []ai.Message) error {
	// Los mensajes nuevos son los que están en after pero no en before
	newMsgs := after[len(before):]
	for _, m := range newMsgs {
		role := entity.AIRoleUser
		if m.Role == ai.RoleAssistant {
			role = entity.AIRoleAssistant
		}
		if err := uc.saveMessage(ctx, convID, role, m.Content); err != nil {
			return err
		}
	}
	return nil
}
