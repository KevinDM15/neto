package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/neto-app/neto/api/internal/domain/valueobject"
)

// TransactionType indica si una transacción es ingreso, egreso o transferencia.
type TransactionType string

const (
	// Income representa un ingreso de dinero a la cuenta.
	Income TransactionType = "income"
	// Expense representa un egreso de dinero de la cuenta.
	Expense TransactionType = "expense"
	// Transfer representa una transferencia entre cuentas del mismo usuario.
	Transfer TransactionType = "transfer"
)

// Transaction representa un movimiento de dinero en una cuenta.
type Transaction struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	AccountID      uuid.UUID
	CategoryID     *uuid.UUID
	Amount         valueobject.Money
	Type           TransactionType
	Description    string
	IdempotencyKey string
	OccurredAt     time.Time
	CreatedAt      time.Time
}

// NewTransaction crea una nueva Transaction validando sus campos obligatorios.
func NewTransaction(
	userID, accountID uuid.UUID,
	amount valueobject.Money,
	txType TransactionType,
	desc, idempotencyKey string,
) (Transaction, error) {
	if amount.IsZero() {
		return Transaction{}, fmt.Errorf("transaction amount cannot be zero")
	}
	switch txType {
	case Income, Expense, Transfer:
	default:
		return Transaction{}, fmt.Errorf("invalid transaction type: %q", txType)
	}
	if idempotencyKey == "" {
		return Transaction{}, fmt.Errorf("idempotency key cannot be empty")
	}

	now := time.Now().UTC()
	return Transaction{
		ID:             uuid.New(),
		UserID:         userID,
		AccountID:      accountID,
		Amount:         amount,
		Type:           txType,
		Description:    desc,
		IdempotencyKey: idempotencyKey,
		OccurredAt:     now,
		CreatedAt:      now,
	}, nil
}

// zero es un helper interno para decimal.Zero — evita repetición.
func zero() decimal.Decimal {
	return decimal.Zero
}
