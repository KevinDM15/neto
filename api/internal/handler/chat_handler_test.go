package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/neto-app/neto/api/internal/handler"
	"github.com/neto-app/neto/api/internal/usecase"
)

// --- Mock de chatUseCase ---

type mockChatUseCase struct {
	response usecase.ChatResponse
	err      error
}

func (m *mockChatUseCase) Chat(_ context.Context, _ uuid.UUID, _ usecase.ChatRequest) (usecase.ChatResponse, error) {
	return m.response, m.err
}

// --- Helpers ---

func chatBody(t *testing.T, v interface{}) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal chat body: %v", err)
	}
	return bytes.NewBuffer(b)
}

// --- Tests ---

// TestChatHandler_NoAuth_Returns401 verifica que sin autenticación se retorna 401.
func TestChatHandler_NoAuth_Returns401(t *testing.T) {
	uc := &mockChatUseCase{}
	h := handler.NewChatHandler(uc)

	body := chatBody(t, map[string]string{"message": "Hola"})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/chat", body)
	w := httptest.NewRecorder()

	h.Chat(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestChatHandler_TextReply_Returns200 verifica que un mensaje simple retorna reply en texto.
func TestChatHandler_TextReply_Returns200(t *testing.T) {
	convID := uuid.New()
	uc := &mockChatUseCase{
		response: usecase.ChatResponse{
			ConversationID: convID,
			Reply:          "Entendido, anotado el gasto.",
		},
	}
	h := handler.NewChatHandler(uc)

	body := chatBody(t, map[string]string{"message": "Gasté 50k en luz"})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/chat", body)
	r = withUserCtx(r, uuid.New())
	w := httptest.NewRecorder()

	h.Chat(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["reply"] == "" {
		t.Error("expected reply field to be non-empty")
	}
	if resp["conversation_id"] != convID.String() {
		t.Errorf("expected conversation_id %s, got %v", convID, resp["conversation_id"])
	}
}

// TestChatHandler_MutationReturnsPending_Returns200 verifica que una mutación retorna pending_confirmation.
func TestChatHandler_MutationReturnsPending_Returns200(t *testing.T) {
	convID := uuid.New()
	preview, _ := json.Marshal(map[string]string{
		"amount":      "50000",
		"description": "Luz",
	})
	uc := &mockChatUseCase{
		response: usecase.ChatResponse{
			ConversationID: convID,
			PendingConfirmation: &usecase.PendingConfirmationResponse{
				Tool:    "create_transaction",
				Preview: preview,
			},
		},
	}
	h := handler.NewChatHandler(uc)

	body := chatBody(t, map[string]string{"message": "Gasté 50k en luz"})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/chat", body)
	r = withUserCtx(r, uuid.New())
	w := httptest.NewRecorder()

	h.Chat(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["pending_confirmation"] == nil {
		t.Error("expected pending_confirmation to be present")
	}
	pending, ok := resp["pending_confirmation"].(map[string]interface{})
	if !ok {
		t.Fatal("pending_confirmation is not an object")
	}
	if pending["tool"] != "create_transaction" {
		t.Errorf("expected tool=create_transaction, got %v", pending["tool"])
	}
}

// TestChatHandler_ConfirmPending_Returns200 verifica que confirmar una acción pendiente funciona.
func TestChatHandler_ConfirmPending_Returns200(t *testing.T) {
	convID := uuid.New()
	uc := &mockChatUseCase{
		response: usecase.ChatResponse{
			ConversationID: convID,
			Reply:          "¡Listo! Gasté $50,000 en Luz registrado.",
		},
	}
	h := handler.NewChatHandler(uc)

	reqBody := map[string]interface{}{
		"conversation_id": convID.String(),
		"confirm":         true,
		"pending_tool": map[string]interface{}{
			"type":  "tool_use",
			"id":    "toolu_01",
			"name":  "create_transaction",
			"input": map[string]string{"amount": "50000", "description": "Luz"},
		},
	}
	body := chatBody(t, reqBody)
	r := httptest.NewRequest(http.MethodPost, "/api/v1/chat", body)
	r = withUserCtx(r, uuid.New())
	w := httptest.NewRecorder()

	h.Chat(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["reply"] == "" {
		t.Error("expected reply field after confirmation")
	}
	if resp["pending_confirmation"] != nil {
		t.Error("expected no pending_confirmation after confirmation")
	}
}

// TestChatHandler_EmptyMessage_Returns400 verifica que un mensaje vacío sin confirm retorna 400.
func TestChatHandler_EmptyMessage_Returns400(t *testing.T) {
	uc := &mockChatUseCase{}
	h := handler.NewChatHandler(uc)

	body := chatBody(t, map[string]string{"message": ""})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/chat", body)
	r = withUserCtx(r, uuid.New())
	w := httptest.NewRecorder()

	h.Chat(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// TestChatHandler_ConfirmPending_Returns200 verifica que confirmar una acción pendiente funciona.
