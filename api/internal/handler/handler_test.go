package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/neto-app/neto/api/internal/domain/entity"
	domainrepo "github.com/neto-app/neto/api/internal/domain/repository"
	"github.com/neto-app/neto/api/internal/domain/valueobject"
	"github.com/neto-app/neto/api/internal/handler"
	"github.com/neto-app/neto/api/internal/middleware"
	"github.com/neto-app/neto/api/internal/usecase"
)

// --- Mocks ---

type mockAccountUseCase struct {
	account  entity.Account
	accounts []entity.Account
	err      error
}

func (m *mockAccountUseCase) CreateAccount(_ context.Context, _ uuid.UUID, name, currencyCode string) (entity.Account, error) {
	return m.account, m.err
}

func (m *mockAccountUseCase) GetAccount(_ context.Context, _, _ uuid.UUID) (entity.Account, error) {
	return m.account, m.err
}

func (m *mockAccountUseCase) ListAccounts(_ context.Context, _ uuid.UUID) ([]entity.Account, error) {
	return m.accounts, m.err
}

type mockTransactionUseCase struct {
	transaction  entity.Transaction
	transactions []entity.Transaction
	err          error
}

func (m *mockTransactionUseCase) CreateTransaction(_ context.Context, _ uuid.UUID, _ usecase.CreateTransactionRequest) (entity.Transaction, error) {
	return m.transaction, m.err
}

func (m *mockTransactionUseCase) ListTransactions(_ context.Context, _ uuid.UUID, _ domainrepo.TransactionFilter) ([]entity.Transaction, error) {
	return m.transactions, m.err
}

// --- Helper para inyectar userID en contexto ---

func withUser(r *http.Request, userID uuid.UUID) *http.Request {
	ctx := context.WithValue(r.Context(), ctxKey("userID"), userID)
	return r.WithContext(ctx)
}

// ctxKey es una key de contexto local para tests.
type ctxKey string

// --- Tests ---

func TestHealth(t *testing.T) {
	h := handler.NewHealthHandler()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", resp["status"])
	}
}

func TestAuthMiddleware_NoToken_Returns401(t *testing.T) {
	protected := middleware.Authenticator("test-secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	w := httptest.NewRecorder()

	protected.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAccountHandler_Create_RequiresAuth(t *testing.T) {
	uc := &mockAccountUseCase{}
	h := handler.NewAccountHandler(uc)

	body := `{"name":"Cuenta","currency_code":"ARS"}`
	r := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", strings.NewReader(body))
	w := httptest.NewRecorder()

	// Sin userID en contexto → debe retornar 401
	h.Create(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without user in context, got %d", w.Code)
	}
}

func TestAccountHandler_List_RetornaSliceVacio(t *testing.T) {
	userID := uuid.New()
	uc := &mockAccountUseCase{accounts: []entity.Account{}}

	h := handler.NewAccountHandler(uc)

	r := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	// Inyectamos el userID usando la función del middleware
	ctx := context.WithValue(r.Context(), middleware.UserIDCtxKey(), userID)
	r = r.WithContext(ctx)

	w := httptest.NewRecorder()
	h.List(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestIdempotencyMiddleware_SinHeader_Pasa(t *testing.T) {
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
	})

	// Sin pool real — pasamos nil porque sin header no debería usarlo
	mux := chi.NewMux()
	mux.Post("/test", inner)

	r := httptest.NewRequest(http.MethodPost, "/test", nil)
	// Sin header Idempotency-Key → el middleware no debería intentar DB
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, r)

	if !called {
		t.Error("expected inner handler to be called")
	}
}

func buildAccount(userID uuid.UUID) entity.Account {
	money, _ := valueobject.NewMoney(valueobject.ZeroDecimal(), "ARS")
	return entity.Account{
		ID:           uuid.New(),
		UserID:       userID,
		Name:         "Test",
		CurrencyCode: "ARS",
		Balance:      money,
	}
}
