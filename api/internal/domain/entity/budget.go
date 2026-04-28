package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/neto-app/neto/api/internal/domain/valueobject"
)

// BudgetPeriod define la recurrencia de un presupuesto.
type BudgetPeriod string

const (
	// Monthly indica un presupuesto mensual.
	Monthly BudgetPeriod = "monthly"
	// Weekly indica un presupuesto semanal.
	Weekly BudgetPeriod = "weekly"
	// Yearly indica un presupuesto anual.
	Yearly BudgetPeriod = "yearly"
)

// Budget representa un límite de gasto por categoría en un período determinado.
type Budget struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	CategoryID uuid.UUID
	Limit      valueobject.Money
	Period     BudgetPeriod
	StartsAt   time.Time
	EndsAt     time.Time
}

// NewBudget crea un nuevo Budget validando sus campos.
func NewBudget(userID, categoryID uuid.UUID, limit valueobject.Money, period BudgetPeriod, startsAt, endsAt time.Time) (Budget, error) {
	if limit.IsZero() {
		return Budget{}, fmt.Errorf("budget limit cannot be zero")
	}
	switch period {
	case Monthly, Weekly, Yearly:
	default:
		return Budget{}, fmt.Errorf("invalid budget period: %q", period)
	}
	if !endsAt.After(startsAt) {
		return Budget{}, fmt.Errorf("budget endsAt must be after startsAt")
	}
	return Budget{
		ID:         uuid.New(),
		UserID:     userID,
		CategoryID: categoryID,
		Limit:      limit,
		Period:     period,
		StartsAt:   startsAt,
		EndsAt:     endsAt,
	}, nil
}

// RemainingAmount retorna cuánto queda del presupuesto dado lo gastado.
func (b Budget) RemainingAmount(spent valueobject.Money) (valueobject.Money, error) {
	return b.Limit.Subtract(spent)
}

// IsExceeded retorna true si lo gastado supera el límite del presupuesto.
func (b Budget) IsExceeded(spent valueobject.Money) (bool, error) {
	if b.Limit.CurrencyCode != spent.CurrencyCode {
		return false, fmt.Errorf("currency mismatch: budget is %s, spent is %s", b.Limit.CurrencyCode, spent.CurrencyCode)
	}
	return spent.Amount.GreaterThan(b.Limit.Amount), nil
}
