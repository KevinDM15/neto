package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"

	"github.com/neto-app/neto/api/internal/infrastructure/anthropic"
	"github.com/neto-app/neto/api/internal/middleware"
	"github.com/neto-app/neto/api/internal/usecase"
)

// chatUseCase define la interfaz que necesita ChatHandler del use case.
type chatUseCase interface {
	Chat(ctx context.Context, userID uuid.UUID, req usecase.ChatRequest) (usecase.ChatResponse, error)
}

// ChatHandler maneja el endpoint HTTP del agente conversacional.
type ChatHandler struct {
	uc chatUseCase
}

// NewChatHandler crea un nuevo ChatHandler.
func NewChatHandler(uc chatUseCase) *ChatHandler {
	return &ChatHandler{uc: uc}
}

// chatRequest es el body esperado en POST /api/v1/chat.
type chatRequest struct {
	ConversationID *string                 `json:"conversation_id"`
	Message        string                  `json:"message"`
	Confirm        bool                    `json:"confirm"`
	PendingTool    *anthropic.ContentBlock `json:"pending_tool"`
}

// chatResponse es la respuesta del endpoint de chat.
type chatResponse struct {
	ConversationID      string                               `json:"conversation_id"`
	Reply               string                               `json:"reply,omitempty"`
	PendingConfirmation *usecase.PendingConfirmationResponse `json:"pending_confirmation,omitempty"`
}

// Chat procesa un mensaje del usuario y retorna la respuesta del agente.
// POST /api/v1/chat
func (h *ChatHandler) Chat(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req chatRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Message == "" && !req.Confirm {
		writeError(w, http.StatusBadRequest, "message is required")
		return
	}

	ucReq := usecase.ChatRequest{
		Message:     req.Message,
		Confirm:     req.Confirm,
		PendingTool: req.PendingTool,
	}

	if req.ConversationID != nil {
		id, err := uuid.Parse(*req.ConversationID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid conversation_id")
			return
		}
		ucReq.ConversationID = &id
	}

	result, err := h.uc.Chat(r.Context(), userID, ucReq)
	if err != nil {
		log.Printf("handler: chat: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to process message")
		return
	}

	writeJSON(w, http.StatusOK, chatResponse{
		ConversationID:      result.ConversationID.String(),
		Reply:               result.Reply,
		PendingConfirmation: result.PendingConfirmation,
	})
}
