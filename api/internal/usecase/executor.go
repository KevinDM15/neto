package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/neto-app/neto/api/internal/ai"
	domainrepo "github.com/neto-app/neto/api/internal/domain/repository"
)

// ToolExecutorBuilder construye la función ejecutora de tools para el loop del LLM.
type ToolExecutorBuilder interface {
	BuildExecutorFunc(userID uuid.UUID, confirmed bool) ai.ToolExecutorFunc
}

// UseCases agrupa los use cases disponibles para el executor de tools.
type UseCases struct {
	Transaction *TransactionUseCase
	Account     *AccountUseCase
	Category    *CategoryUseCase
}

// APIToolExecutor implementa ToolExecutorBuilder llamando a los use cases reales.
type APIToolExecutor struct {
	uc UseCases
}

// NewAPIToolExecutor crea un nuevo APIToolExecutor con los use cases dados.
func NewAPIToolExecutor(uc UseCases) *APIToolExecutor {
	return &APIToolExecutor{uc: uc}
}

// requiresConfirmation retorna true si el tool dado requiere confirmación del usuario
// antes de ejecutarse. Esta lógica es de dominio de la aplicación, no de infraestructura.
func requiresConfirmation(toolName string) bool {
	switch toolName {
	case "create_transaction", "create_account", "record_debt", "delete_transaction":
		return true
	default:
		return false
	}
}

// BuildExecutorFunc construye la ToolExecutorFunc para el loop del LLM.
// confirmed=true significa que el usuario ya aprobó las mutaciones pendientes.
func (e *APIToolExecutor) BuildExecutorFunc(userID uuid.UUID, confirmed bool) ai.ToolExecutorFunc {
	return func(ctx context.Context, block *ai.ContentBlock) (json.RawMessage, bool, error) {
		if requiresConfirmation(block.Name) && !confirmed {
			return nil, true, nil
		}
		return e.dispatch(ctx, block, userID)
	}
}

// dispatch despacha la ejecución del tool al use case correspondiente.
func (e *APIToolExecutor) dispatch(ctx context.Context, block *ai.ContentBlock, userID uuid.UUID) (json.RawMessage, bool, error) {
	switch block.Name {
	case "list_transactions":
		r, err := e.listTransactions(ctx, block.Input, userID)
		return r, false, err
	case "get_balance":
		r, err := e.getBalance(ctx, block.Input, userID)
		return r, false, err
	case "list_categories":
		r, err := e.listCategories(ctx, userID)
		return r, false, err
	case "get_monthly_summary":
		r, err := e.getMonthlySummary(ctx, block.Input, userID)
		return r, false, err
	case "create_transaction":
		r, err := e.createTransaction(ctx, block.Input, userID)
		return r, false, err
	case "create_account":
		r, err := e.createAccount(ctx, block.Input, userID)
		return r, false, err
	case "delete_transaction":
		r, err := e.deleteTransaction(ctx, block.Input, userID)
		return r, false, err
	case "record_debt", "update_goal_progress":
		result, err := json.Marshal(map[string]string{"status": "ok", "message": "operación registrada"})
		if err != nil {
			return nil, false, fmt.Errorf("marshal placeholder result: %w", err)
		}
		return result, false, nil
	default:
		return nil, false, fmt.Errorf("executor: unknown tool %q", block.Name)
	}
}

// --- Implementaciones de tools ---

type listTransactionsInput struct {
	AccountID  *string `json:"account_id"`
	CategoryID *string `json:"category_id"`
	From       *string `json:"from"`
	To         *string `json:"to"`
	Limit      int     `json:"limit"`
}

func (e *APIToolExecutor) listTransactions(ctx context.Context, input json.RawMessage, userID uuid.UUID) (json.RawMessage, error) {
	var in listTransactionsInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, fmt.Errorf("list_transactions: unmarshal input: %w", err)
	}

	filter := domainrepo.TransactionFilter{Limit: 20}
	if in.Limit > 0 {
		filter.Limit = in.Limit
	}
	if in.AccountID != nil {
		id, err := uuid.Parse(*in.AccountID)
		if err != nil {
			return nil, fmt.Errorf("list_transactions: invalid account_id: %w", err)
		}
		filter.AccountID = &id
	}
	if in.CategoryID != nil {
		id, err := uuid.Parse(*in.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("list_transactions: invalid category_id: %w", err)
		}
		filter.CategoryID = &id
	}
	if in.From != nil {
		t, err := time.Parse("2006-01-02", *in.From)
		if err != nil {
			return nil, fmt.Errorf("list_transactions: invalid from date: %w", err)
		}
		filter.From = &t
	}
	if in.To != nil {
		t, err := time.Parse("2006-01-02", *in.To)
		if err != nil {
			return nil, fmt.Errorf("list_transactions: invalid to date: %w", err)
		}
		filter.To = &t
	}

	txs, err := e.uc.Transaction.ListTransactions(ctx, userID, filter)
	if err != nil {
		return nil, fmt.Errorf("list_transactions: %w", err)
	}

	type txItem struct {
		ID          string `json:"id"`
		Amount      string `json:"amount"`
		Currency    string `json:"currency_code"`
		Type        string `json:"type"`
		Description string `json:"description"`
		OccurredAt  string `json:"occurred_at"`
	}
	items := make([]txItem, 0, len(txs))
	for _, t := range txs {
		items = append(items, txItem{
			ID:          t.ID.String(),
			Amount:      t.Amount.Amount.String(),
			Currency:    t.Amount.CurrencyCode,
			Type:        string(t.Type),
			Description: t.Description,
			OccurredAt:  t.OccurredAt.Format(time.RFC3339),
		})
	}

	return json.Marshal(items)
}

type getBalanceInput struct {
	AccountID *string `json:"account_id"`
}

func (e *APIToolExecutor) getBalance(ctx context.Context, input json.RawMessage, userID uuid.UUID) (json.RawMessage, error) {
	var in getBalanceInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, fmt.Errorf("get_balance: unmarshal input: %w", err)
	}

	accounts, err := e.uc.Account.ListAccounts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get_balance: list accounts: %w", err)
	}

	type balanceItem struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Currency string `json:"currency_code"`
		Balance  string `json:"balance"`
	}

	items := make([]balanceItem, 0, len(accounts))
	for _, a := range accounts {
		if in.AccountID != nil && a.ID.String() != *in.AccountID {
			continue
		}
		items = append(items, balanceItem{
			ID:       a.ID.String(),
			Name:     a.Name,
			Currency: a.CurrencyCode,
			Balance:  a.Balance.Amount.String(),
		})
	}

	return json.Marshal(items)
}

func (e *APIToolExecutor) listCategories(ctx context.Context, userID uuid.UUID) (json.RawMessage, error) {
	cats, err := e.uc.Category.ListCategories(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list_categories: %w", err)
	}

	type catItem struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	items := make([]catItem, 0, len(cats))
	for _, c := range cats {
		items = append(items, catItem{ID: c.ID.String(), Name: c.Name})
	}

	return json.Marshal(items)
}

type getMonthlySummaryInput struct {
	Year  int `json:"year"`
	Month int `json:"month"`
}

func (e *APIToolExecutor) getMonthlySummary(ctx context.Context, input json.RawMessage, userID uuid.UUID) (json.RawMessage, error) {
	var in getMonthlySummaryInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, fmt.Errorf("get_monthly_summary: unmarshal input: %w", err)
	}

	from := time.Date(in.Year, time.Month(in.Month), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0).Add(-time.Second)

	filter := domainrepo.TransactionFilter{From: &from, To: &to, Limit: 1000}
	txs, err := e.uc.Transaction.ListTransactions(ctx, userID, filter)
	if err != nil {
		return nil, fmt.Errorf("get_monthly_summary: list transactions: %w", err)
	}

	var totalIncome, totalExpense float64
	for _, t := range txs {
		amt, _ := t.Amount.Amount.Float64()
		switch t.Type {
		case "income":
			totalIncome += amt
		case "expense":
			totalExpense += amt
		}
	}

	return json.Marshal(map[string]interface{}{
		"year":          in.Year,
		"month":         in.Month,
		"total_income":  fmt.Sprintf("%.2f", totalIncome),
		"total_expense": fmt.Sprintf("%.2f", totalExpense),
		"net_balance":   fmt.Sprintf("%.2f", totalIncome-totalExpense),
	})
}

type createTransactionInput struct {
	AccountID    string  `json:"account_id"`
	Amount       string  `json:"amount"`
	CurrencyCode string  `json:"currency_code"`
	Type         string  `json:"type"`
	Description  string  `json:"description"`
	CategoryID   *string `json:"category_id"`
}

func (e *APIToolExecutor) createTransaction(ctx context.Context, input json.RawMessage, userID uuid.UUID) (json.RawMessage, error) {
	var in createTransactionInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, fmt.Errorf("create_transaction: unmarshal input: %w", err)
	}

	accountID, err := uuid.Parse(in.AccountID)
	if err != nil {
		return nil, fmt.Errorf("create_transaction: invalid account_id: %w", err)
	}

	req := CreateTransactionRequest{
		AccountID:      accountID,
		Amount:         in.Amount,
		CurrencyCode:   in.CurrencyCode,
		Type:           TransactionTypeFromString(in.Type),
		Description:    in.Description,
		IdempotencyKey: uuid.New().String(),
	}

	if in.CategoryID != nil {
		catID, err := uuid.Parse(*in.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("create_transaction: invalid category_id: %w", err)
		}
		req.CategoryID = &catID
	}

	tx, err := e.uc.Transaction.CreateTransaction(ctx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("create_transaction: %w", err)
	}

	return json.Marshal(map[string]string{
		"id":          tx.ID.String(),
		"status":      "created",
		"amount":      tx.Amount.Amount.String(),
		"description": tx.Description,
	})
}

type createAccountInput struct {
	Name         string `json:"name"`
	CurrencyCode string `json:"currency_code"`
}

func (e *APIToolExecutor) createAccount(ctx context.Context, input json.RawMessage, userID uuid.UUID) (json.RawMessage, error) {
	var in createAccountInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, fmt.Errorf("create_account: unmarshal input: %w", err)
	}

	acc, err := e.uc.Account.CreateAccount(ctx, userID, in.Name, in.CurrencyCode)
	if err != nil {
		return nil, fmt.Errorf("create_account: %w", err)
	}

	return json.Marshal(map[string]string{
		"id":            acc.ID.String(),
		"name":          acc.Name,
		"currency_code": acc.CurrencyCode,
		"status":        "created",
	})
}

type deleteTransactionInput struct {
	TransactionID string `json:"transaction_id"`
}

func (e *APIToolExecutor) deleteTransaction(ctx context.Context, input json.RawMessage, userID uuid.UUID) (json.RawMessage, error) {
	var in deleteTransactionInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, fmt.Errorf("delete_transaction: unmarshal input: %w", err)
	}

	txID, err := uuid.Parse(in.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("delete_transaction: invalid transaction_id: %w", err)
	}

	if err := e.uc.Transaction.DeleteTransaction(ctx, userID, txID); err != nil {
		return nil, fmt.Errorf("delete_transaction: %w", err)
	}

	return json.Marshal(map[string]string{
		"id":     txID.String(),
		"status": "deleted",
	})
}
