package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/neto-app/neto/tui/internal/client"
	"github.com/neto-app/neto/tui/internal/config"
)

// newTestClient builds a Client wired to both a mock API server and a mock
// Supabase server. Pass "" for supabaseServer to use the default Supabase URL.
func newTestClient(apiServer, supabaseServer string) *client.Client {
	cfg := &config.Config{
		APIURL:          apiServer,
		SupabaseURL:     supabaseServer,
		SupabaseAnonKey: "test-anon-key",
	}
	return client.New(cfg, nil)
}

// ── Login ──────────────────────────────────────────────────────────────────

func TestClient_Login_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/v1/token" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "at-123",
			"refresh_token": "rt-456",
			"expires_in":    3600,
		})
	}))
	defer srv.Close()

	c := newTestClient("", srv.URL)
	at, rt, err := c.Login(context.Background(), "user@example.com", "secret")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if at != "at-123" {
		t.Errorf("access_token: got %q, want %q", at, "at-123")
	}
	if rt != "rt-456" {
		t.Errorf("refresh_token: got %q, want %q", rt, "rt-456")
	}
}

func TestClient_Login_Failure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "Invalid credentials",
		})
	}))
	defer srv.Close()

	c := newTestClient("", srv.URL)
	_, _, err := c.Login(context.Background(), "bad@example.com", "wrong")
	if err == nil {
		t.Fatal("expected error on 400, got nil")
	}
}

// ── Chat ──────────────────────────────────────────────────────────────────

func TestClient_Chat_TextReply(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/chat" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"conversation_id": "conv-1",
			"reply":           "Hola, ¿en qué puedo ayudarte?",
		})
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	resp, err := c.Chat(context.Background(), client.ChatRequest{Message: "hola"})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if resp.Reply != "Hola, ¿en qué puedo ayudarte?" {
		t.Errorf("reply: got %q", resp.Reply)
	}
	if resp.PendingConfirmation != nil {
		t.Error("expected no pending_confirmation")
	}
}

func TestClient_Chat_PendingConfirmation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/chat" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"conversation_id": "conv-2",
			"reply":           "¿Confirmo el gasto de $50,000?",
			"pending_confirmation": map[string]interface{}{
				"tool": "create_transaction",
				"preview": map[string]interface{}{
					"amount":      50000,
					"description": "Luz",
				},
			},
		})
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	resp, err := c.Chat(context.Background(), client.ChatRequest{Message: "Gasté 50k en luz"})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if resp.PendingConfirmation == nil {
		t.Fatal("expected pending_confirmation, got nil")
	}
	if resp.PendingConfirmation.Tool != "create_transaction" {
		t.Errorf("tool: got %q, want %q", resp.PendingConfirmation.Tool, "create_transaction")
	}
}
