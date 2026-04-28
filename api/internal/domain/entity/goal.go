package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/neto-app/neto/api/internal/domain/valueobject"
)

// Goal representa una meta de ahorro del usuario.
type Goal struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Name          string
	TargetAmount  valueobject.Money
	CurrentAmount valueobject.Money
	Deadline      *time.Time
}

// NewGoal crea una nueva Goal validando sus campos obligatorios.
func NewGoal(userID uuid.UUID, name string, targetAmount valueobject.Money) (Goal, error) {
	if name == "" {
		return Goal{}, fmt.Errorf("goal name cannot be empty")
	}
	if targetAmount.IsZero() {
		return Goal{}, fmt.Errorf("goal target amount cannot be zero")
	}
	zeroBalance, err := valueobject.NewMoney(decimal.Zero, targetAmount.CurrencyCode)
	if err != nil {
		return Goal{}, fmt.Errorf("failed to create zero current amount: %w", err)
	}
	return Goal{
		ID:            uuid.New(),
		UserID:        userID,
		Name:          name,
		TargetAmount:  targetAmount,
		CurrentAmount: zeroBalance,
	}, nil
}

// Progress retorna el porcentaje de avance hacia la meta (0-100).
func (g Goal) Progress() float64 {
	if g.TargetAmount.IsZero() {
		return 0
	}
	pct, _ := g.CurrentAmount.Amount.Div(g.TargetAmount.Amount).Mul(decimal.NewFromInt(100)).Float64()
	if pct > 100 {
		return 100
	}
	return pct
}

// IsCompleted retorna true si el monto actual alcanzó o superó la meta.
func (g Goal) IsCompleted() bool {
	return g.CurrentAmount.Amount.GreaterThanOrEqual(g.TargetAmount.Amount)
}

// RemainingAmount retorna cuánto falta para alcanzar la meta.
func (g Goal) RemainingAmount() (valueobject.Money, error) {
	return g.TargetAmount.Subtract(g.CurrentAmount)
}
