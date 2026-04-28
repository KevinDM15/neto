package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neto-app/neto/api/internal/domain/valueobject"
)

// DebtDirection indica si la deuda es a favor o en contra del usuario.
type DebtDirection string

const (
	// Owed indica que alguien le debe dinero al usuario.
	Owed DebtDirection = "owed"
	// Owing indica que el usuario le debe dinero a alguien.
	Owing DebtDirection = "owing"
)

// Debt representa una deuda entre el usuario y una contraparte.
type Debt struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	CounterpartyName string
	Amount           valueobject.Money
	Direction        DebtDirection
	DueDate          *time.Time
	PaidAt           *time.Time
	Notes            string
}

// NewDebt crea una nueva Debt validando sus campos obligatorios.
func NewDebt(userID uuid.UUID, counterpartyName string, amount valueobject.Money, direction DebtDirection) (Debt, error) {
	if counterpartyName == "" {
		return Debt{}, fmt.Errorf("counterparty name cannot be empty")
	}
	if amount.IsZero() {
		return Debt{}, fmt.Errorf("debt amount cannot be zero")
	}
	switch direction {
	case Owed, Owing:
	default:
		return Debt{}, fmt.Errorf("invalid debt direction: %q", direction)
	}
	return Debt{
		ID:               uuid.New(),
		UserID:           userID,
		CounterpartyName: counterpartyName,
		Amount:           amount,
		Direction:        direction,
	}, nil
}

// IsPaid retorna true si la deuda fue saldada.
func (d Debt) IsPaid() bool {
	return d.PaidAt != nil
}

// IsOverdue retorna true si la deuda tiene fecha de vencimiento, no fue pagada y ya venció.
func (d Debt) IsOverdue() bool {
	if d.IsPaid() || d.DueDate == nil {
		return false
	}
	return time.Now().UTC().After(*d.DueDate)
}
