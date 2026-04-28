package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/neto-app/neto/api/internal/domain/entity"
	domainrepo "github.com/neto-app/neto/api/internal/domain/repository"
	"github.com/neto-app/neto/api/internal/domain/valueobject"
)

// CreateTransactionRequest agrupa los datos necesarios para crear una transacción.
type CreateTransactionRequest struct {
	AccountID      uuid.UUID
	CategoryID     *uuid.UUID
	Amount         string // string decimal para evitar float64
	CurrencyCode   string
	Type           entity.TransactionType
	Description    string
	IdempotencyKey string
}

// TransactionUseCase contiene la lógica de aplicación para transacciones.
type TransactionUseCase struct {
	transactions domainrepo.TransactionRepository
	accounts     domainrepo.AccountRepository
}

// NewTransactionUseCase crea un nuevo TransactionUseCase.
func NewTransactionUseCase(transactions domainrepo.TransactionRepository, accounts domainrepo.AccountRepository) *TransactionUseCase {
	return &TransactionUseCase{transactions: transactions, accounts: accounts}
}

// CreateTransaction crea y persiste una nueva transacción.
func (uc *TransactionUseCase) CreateTransaction(ctx context.Context, userID uuid.UUID, req CreateTransactionRequest) (entity.Transaction, error) {
	// Validar que la cuenta pertenezca al usuario
	account, err := uc.accounts.GetByID(ctx, req.AccountID)
	if err != nil {
		return entity.Transaction{}, fmt.Errorf("usecase: get account for transaction: %w", err)
	}
	if account.UserID != userID {
		return entity.Transaction{}, fmt.Errorf("usecase: account does not belong to user")
	}

	// Parsear el monto — NUNCA float64
	amountDecimal, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return entity.Transaction{}, fmt.Errorf("usecase: invalid amount %q: %w", req.Amount, err)
	}

	currencyCode := req.CurrencyCode
	if currencyCode == "" {
		currencyCode = account.CurrencyCode
	}

	money, err := valueobject.NewMoney(amountDecimal, currencyCode)
	if err != nil {
		return entity.Transaction{}, fmt.Errorf("usecase: build money: %w", err)
	}

	t, err := entity.NewTransaction(userID, req.AccountID, money, req.Type, req.Description, req.IdempotencyKey)
	if err != nil {
		return entity.Transaction{}, fmt.Errorf("usecase: create transaction entity: %w", err)
	}
	t.CategoryID = req.CategoryID

	if err := uc.transactions.Create(ctx, t); err != nil {
		return entity.Transaction{}, fmt.Errorf("usecase: persist transaction: %w", err)
	}

	return t, nil
}

// ListTransactions retorna transacciones del usuario con filtros opcionales.
func (uc *TransactionUseCase) ListTransactions(ctx context.Context, userID uuid.UUID, filter domainrepo.TransactionFilter) ([]entity.Transaction, error) {
	transactions, err := uc.transactions.GetByUserID(ctx, userID, filter)
	if err != nil {
		return nil, fmt.Errorf("usecase: list transactions: %w", err)
	}
	return transactions, nil
}

// DeleteTransaction elimina una transacción verificando que pertenezca al usuario.
func (uc *TransactionUseCase) DeleteTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) error {
	tx, err := uc.transactions.GetByID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("usecase: get transaction for delete: %w", err)
	}
	if tx.UserID != userID {
		return fmt.Errorf("usecase: transaction does not belong to user")
	}
	if err := uc.transactions.Delete(ctx, transactionID); err != nil {
		return fmt.Errorf("usecase: delete transaction: %w", err)
	}
	return nil
}

// TransactionTypeFromString convierte un string al tipo TransactionType correspondiente.
func TransactionTypeFromString(s string) entity.TransactionType {
	switch s {
	case "income":
		return entity.Income
	case "transfer":
		return entity.Transfer
	default:
		return entity.Expense
	}
}
