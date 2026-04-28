// Package usecase implementa la lógica de aplicación de Neto.
// Orquesta el dominio y los repositorios sin conocer HTTP ni DB directamente.
package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/neto-app/neto/api/internal/domain/entity"
	domainrepo "github.com/neto-app/neto/api/internal/domain/repository"
)

// AccountUseCase contiene la lógica de aplicación para cuentas.
type AccountUseCase struct {
	accounts   domainrepo.AccountRepository
	currencies domainrepo.CurrencyRepository
}

// NewAccountUseCase crea un nuevo AccountUseCase.
func NewAccountUseCase(accounts domainrepo.AccountRepository, currencies domainrepo.CurrencyRepository) *AccountUseCase {
	return &AccountUseCase{accounts: accounts, currencies: currencies}
}

// CreateAccount crea una nueva cuenta para el usuario después de validar la moneda.
func (uc *AccountUseCase) CreateAccount(ctx context.Context, userID uuid.UUID, name, currencyCode string) (entity.Account, error) {
	// Validar que la moneda exista y esté activa
	if _, err := uc.currencies.GetByCode(ctx, currencyCode); err != nil {
		return entity.Account{}, fmt.Errorf("usecase: invalid currency %q: %w", currencyCode, err)
	}

	account, err := entity.NewAccount(userID, name, currencyCode)
	if err != nil {
		return entity.Account{}, fmt.Errorf("usecase: create account entity: %w", err)
	}

	if err := uc.accounts.Create(ctx, account); err != nil {
		return entity.Account{}, fmt.Errorf("usecase: persist account: %w", err)
	}

	return account, nil
}

// GetAccount retorna una cuenta verificando que pertenezca al usuario.
func (uc *AccountUseCase) GetAccount(ctx context.Context, userID, accountID uuid.UUID) (entity.Account, error) {
	account, err := uc.accounts.GetByID(ctx, accountID)
	if err != nil {
		return entity.Account{}, fmt.Errorf("usecase: get account: %w", err)
	}

	// Doble check de propiedad (además del RLS en DB)
	if account.UserID != userID {
		return entity.Account{}, fmt.Errorf("usecase: account does not belong to user")
	}

	return account, nil
}

// ListAccounts retorna todas las cuentas del usuario.
func (uc *AccountUseCase) ListAccounts(ctx context.Context, userID uuid.UUID) ([]entity.Account, error) {
	accounts, err := uc.accounts.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("usecase: list accounts: %w", err)
	}
	return accounts, nil
}
