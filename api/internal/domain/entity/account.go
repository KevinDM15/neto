// Package entity contiene las entidades del dominio de Neto.
// Las entidades tienen identidad propia y ciclo de vida.
package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neto-app/neto/api/internal/domain/valueobject"
)

// Account representa una cuenta financiera de un usuario (ej: cuenta bancaria, billetera virtual).
type Account struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Name         string
	CurrencyCode string
	Balance      valueobject.Money
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewAccount crea una nueva Account con balance inicial en cero.
func NewAccount(userID uuid.UUID, name, currencyCode string) (Account, error) {
	if name == "" {
		return Account{}, fmt.Errorf("account name cannot be empty")
	}
	if currencyCode == "" {
		return Account{}, fmt.Errorf("currency code cannot be empty")
	}

	now := time.Now().UTC()
	zeroBalance, err := valueobject.NewMoney(zero(), currencyCode)
	if err != nil {
		return Account{}, fmt.Errorf("failed to create zero balance: %w", err)
	}

	return Account{
		ID:           uuid.New(),
		UserID:       userID,
		Name:         name,
		CurrencyCode: currencyCode,
		Balance:      zeroBalance,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// ApplyTransaction actualiza el balance de la cuenta según el tipo de transacción.
func (a *Account) ApplyTransaction(t Transaction) error {
	switch t.Type {
	case Income:
		newBalance, err := a.Balance.Add(t.Amount)
		if err != nil {
			return fmt.Errorf("applying income transaction: %w", err)
		}
		a.Balance = newBalance
	case Expense:
		newBalance, err := a.Balance.Subtract(t.Amount)
		if err != nil {
			return fmt.Errorf("applying expense transaction: %w", err)
		}
		a.Balance = newBalance
	case Transfer:
		newBalance, err := a.Balance.Subtract(t.Amount)
		if err != nil {
			return fmt.Errorf("applying transfer transaction: %w", err)
		}
		a.Balance = newBalance
	default:
		return fmt.Errorf("unknown transaction type: %s", t.Type)
	}
	a.UpdatedAt = time.Now().UTC()
	return nil
}
