package handler

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/neto-app/neto/api/internal/domain/entity"
	domainrepo "github.com/neto-app/neto/api/internal/domain/repository"
	"github.com/neto-app/neto/api/internal/middleware"
	"github.com/neto-app/neto/api/internal/usecase"
)

// transactionUseCase define la interfaz que necesita TransactionHandler.
type transactionUseCase interface {
	CreateTransaction(ctx context.Context, userID uuid.UUID, req usecase.CreateTransactionRequest) (entity.Transaction, error)
	ListTransactions(ctx context.Context, userID uuid.UUID, filter domainrepo.TransactionFilter) ([]entity.Transaction, error)
}

// TransactionHandler maneja los endpoints HTTP de transacciones.
type TransactionHandler struct {
	uc transactionUseCase
}

// NewTransactionHandler crea un nuevo TransactionHandler.
func NewTransactionHandler(uc transactionUseCase) *TransactionHandler {
	return &TransactionHandler{uc: uc}
}

// createTransactionRequest es el body esperado para crear una transacción.
type createTransactionRequest struct {
	AccountID      string  `json:"account_id"`
	CategoryID     *string `json:"category_id,omitempty"`
	Amount         string  `json:"amount"`
	CurrencyCode   string  `json:"currency_code,omitempty"`
	Type           string  `json:"type"`
	Description    string  `json:"description,omitempty"`
	IdempotencyKey string  `json:"idempotency_key"`
}

// transactionResponse es la representación JSON de una transacción.
type transactionResponse struct {
	ID             string  `json:"id"`
	AccountID      string  `json:"account_id"`
	CategoryID     *string `json:"category_id,omitempty"`
	Amount         string  `json:"amount"`
	Type           string  `json:"type"`
	Description    string  `json:"description"`
	IdempotencyKey string  `json:"idempotency_key"`
	OccurredAt     string  `json:"occurred_at"`
	CreatedAt      string  `json:"created_at"`
}

func toTransactionResponse(t entity.Transaction) transactionResponse {
	resp := transactionResponse{
		ID:             t.ID.String(),
		AccountID:      t.AccountID.String(),
		Amount:         t.Amount.Amount.String(),
		Type:           string(t.Type),
		Description:    t.Description,
		IdempotencyKey: t.IdempotencyKey,
		OccurredAt:     t.OccurredAt.Format("2006-01-02T15:04:05Z"),
		CreatedAt:      t.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if t.CategoryID != nil {
		s := t.CategoryID.String()
		resp.CategoryID = &s
	}
	return resp
}

// Create crea una nueva transacción para el usuario autenticado.
// POST /api/v1/transactions
func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createTransactionRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.AccountID == "" || req.Amount == "" || req.Type == "" || req.IdempotencyKey == "" {
		writeError(w, http.StatusBadRequest, "account_id, amount, type and idempotency_key are required")
		return
	}

	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid account_id")
		return
	}

	ucReq := usecase.CreateTransactionRequest{
		AccountID:      accountID,
		Amount:         req.Amount,
		CurrencyCode:   req.CurrencyCode,
		Type:           entity.TransactionType(req.Type),
		Description:    req.Description,
		IdempotencyKey: req.IdempotencyKey,
	}

	if req.CategoryID != nil {
		catID, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid category_id")
			return
		}
		ucReq.CategoryID = &catID
	}

	t, err := h.uc.CreateTransaction(r.Context(), userID, ucReq)
	if err != nil {
		log.Printf("handler: create transaction: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create transaction")
		return
	}

	writeJSON(w, http.StatusCreated, toTransactionResponse(t))
}

// List retorna las transacciones del usuario con filtros opcionales vía query params.
// GET /api/v1/transactions
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	filter := domainrepo.TransactionFilter{}

	if accountIDStr := r.URL.Query().Get("account_id"); accountIDStr != "" {
		id, err := uuid.Parse(accountIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid account_id filter")
			return
		}
		filter.AccountID = &id
	}

	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		t, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid from date, use YYYY-MM-DD")
			return
		}
		filter.From = &t
	}

	if toStr := r.URL.Query().Get("to"); toStr != "" {
		t, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid to date, use YYYY-MM-DD")
			return
		}
		filter.To = &t
	}

	transactions, err := h.uc.ListTransactions(r.Context(), userID, filter)
	if err != nil {
		log.Printf("handler: list transactions: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list transactions")
		return
	}

	resp := make([]transactionResponse, 0, len(transactions))
	for _, t := range transactions {
		resp = append(resp, toTransactionResponse(t))
	}

	writeJSON(w, http.StatusOK, resp)
}
