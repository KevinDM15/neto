// Package repository define las interfaces de persistencia del dominio de Neto.
// Las implementaciones concretas viven en api/internal/repository/ (capa de infraestructura).
// Esta separación sigue el principio de Inversión de Dependencias de Clean Architecture:
// el dominio define QUÉ necesita, la infraestructura define CÓMO lo provee.
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/neto-app/neto/api/internal/domain/entity"
	"github.com/neto-app/neto/api/internal/domain/valueobject"
)

// AccountRepository define las operaciones de persistencia para cuentas.
type AccountRepository interface {
	Create(ctx context.Context, account entity.Account) error
	GetByID(ctx context.Context, id uuid.UUID) (entity.Account, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Account, error)
	Update(ctx context.Context, account entity.Account) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// CategoryRepository define las operaciones de persistencia para categorías.
type CategoryRepository interface {
	Create(ctx context.Context, category entity.Category) error
	GetByID(ctx context.Context, id uuid.UUID) (entity.Category, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Category, error)
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]entity.Category, error)
	Update(ctx context.Context, category entity.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// TransactionFilter agrupa los filtros opcionales para listar transacciones.
type TransactionFilter struct {
	AccountID  *uuid.UUID
	CategoryID *uuid.UUID
	From       *time.Time
	To         *time.Time
	Limit      int
	Offset     int
}

// TransactionRepository define las operaciones de persistencia para transacciones.
type TransactionRepository interface {
	Create(ctx context.Context, transaction entity.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (entity.Transaction, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, filter TransactionFilter) ([]entity.Transaction, error)
	GetByIdempotencyKey(ctx context.Context, key string) (entity.Transaction, error)
	Update(ctx context.Context, transaction entity.Transaction) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// BudgetRepository define las operaciones de persistencia para presupuestos.
type BudgetRepository interface {
	Create(ctx context.Context, budget entity.Budget) error
	GetByID(ctx context.Context, id uuid.UUID) (entity.Budget, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Budget, error)
	GetActive(ctx context.Context, userID uuid.UUID, date time.Time) ([]entity.Budget, error)
	Update(ctx context.Context, budget entity.Budget) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// DebtRepository define las operaciones de persistencia para deudas.
type DebtRepository interface {
	Create(ctx context.Context, debt entity.Debt) error
	GetByID(ctx context.Context, id uuid.UUID) (entity.Debt, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Debt, error)
	GetUnpaid(ctx context.Context, userID uuid.UUID) ([]entity.Debt, error)
	Update(ctx context.Context, debt entity.Debt) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// GoalRepository define las operaciones de persistencia para metas de ahorro.
type GoalRepository interface {
	Create(ctx context.Context, goal entity.Goal) error
	GetByID(ctx context.Context, id uuid.UUID) (entity.Goal, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Goal, error)
	Update(ctx context.Context, goal entity.Goal) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// CurrencyRepository define las operaciones de persistencia para monedas.
type CurrencyRepository interface {
	GetAll(ctx context.Context) ([]valueobject.Currency, error)
	GetByCode(ctx context.Context, code string) (valueobject.Currency, error)
	GetActive(ctx context.Context) ([]valueobject.Currency, error)
}
