package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/neto-app/neto/api/internal/domain/entity"
	"github.com/neto-app/neto/api/internal/middleware"
)

// accountUseCase define la interfaz que necesita AccountHandler del use case.
type accountUseCase interface {
	CreateAccount(ctx context.Context, userID uuid.UUID, name, currencyCode string) (entity.Account, error)
	GetAccount(ctx context.Context, userID, accountID uuid.UUID) (entity.Account, error)
	ListAccounts(ctx context.Context, userID uuid.UUID) ([]entity.Account, error)
}

// AccountHandler maneja los endpoints HTTP de cuentas.
type AccountHandler struct {
	uc accountUseCase
}

// NewAccountHandler crea un nuevo AccountHandler.
func NewAccountHandler(uc accountUseCase) *AccountHandler {
	return &AccountHandler{uc: uc}
}

// createAccountRequest es el body esperado para crear una cuenta.
type createAccountRequest struct {
	Name         string `json:"name"`
	CurrencyCode string `json:"currency_code"`
}

// accountResponse es la representación JSON de una cuenta.
type accountResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	CurrencyCode string `json:"currency_code"`
	Balance      string `json:"balance"`
	CreatedAt    string `json:"created_at"`
}

// toAccountResponse convierte un entity.Account a su representación JSON.
func toAccountResponse(a entity.Account) accountResponse {
	return accountResponse{
		ID:           a.ID.String(),
		Name:         a.Name,
		CurrencyCode: a.CurrencyCode,
		Balance:      a.Balance.Amount.String(),
		CreatedAt:    a.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// Create crea una nueva cuenta para el usuario autenticado.
// POST /api/v1/accounts
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createAccountRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.CurrencyCode == "" {
		writeError(w, http.StatusBadRequest, "name and currency_code are required")
		return
	}

	account, err := h.uc.CreateAccount(r.Context(), userID, req.Name, req.CurrencyCode)
	if err != nil {
		log.Printf("handler: create account: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create account")
		return
	}

	writeJSON(w, http.StatusCreated, toAccountResponse(account))
}

// List retorna todas las cuentas del usuario autenticado.
// GET /api/v1/accounts
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	accounts, err := h.uc.ListAccounts(r.Context(), userID)
	if err != nil {
		log.Printf("handler: list accounts: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list accounts")
		return
	}

	resp := make([]accountResponse, 0, len(accounts))
	for _, a := range accounts {
		resp = append(resp, toAccountResponse(a))
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetByID retorna el detalle de una cuenta por ID.
// GET /api/v1/accounts/{id}
func (h *AccountHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	accountID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid account id")
		return
	}

	account, err := h.uc.GetAccount(r.Context(), userID, accountID)
	if err != nil {
		log.Printf("handler: get account %s: %v", accountID, err)
		writeError(w, http.StatusNotFound, "account not found")
		return
	}

	writeJSON(w, http.StatusOK, toAccountResponse(account))
}
