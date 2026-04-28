// Package repository implementa las interfaces de dominio usando Supabase (pgx).
// Cada tipo en este paquete corresponde a una interface del domain/repository.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/neto-app/neto/api/internal/domain/entity"
	"github.com/neto-app/neto/api/internal/domain/valueobject"
)

// AccountRepository implementa domain/repository.AccountRepository usando pgx.
type AccountRepository struct {
	pool *pgxpool.Pool
}

// newAccountRepository crea un AccountRepository con el pool dado.
func newAccountRepository(pool *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{pool: pool}
}

// Create inserta una nueva cuenta en la DB.
func (r *AccountRepository) Create(ctx context.Context, account entity.Account) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO accounts (id, user_id, name, currency_code, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		account.ID,
		account.UserID,
		account.Name,
		account.CurrencyCode,
		account.Balance.Amount.String(),
		account.CreatedAt,
		account.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository: create account: %w", err)
	}
	return nil
}

// GetByID busca una cuenta por ID. Incluye filtro por user_id como doble check además del RLS.
func (r *AccountRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.Account, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, name, currency_code, balance, created_at, updated_at
		FROM accounts
		WHERE id = $1
	`, id)

	return scanAccount(row)
}

// GetByUserID retorna todas las cuentas del usuario.
func (r *AccountRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Account, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, name, currency_code, balance, created_at, updated_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: get accounts by user: %w", err)
	}
	defer rows.Close()

	var accounts []entity.Account
	for rows.Next() {
		acc, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, rows.Err()
}

// Update actualiza balance y updated_at de una cuenta.
func (r *AccountRepository) Update(ctx context.Context, account entity.Account) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE accounts
		SET balance = $1, name = $2, updated_at = $3
		WHERE id = $4 AND user_id = $5
	`,
		account.Balance.Amount.String(),
		account.Name,
		time.Now().UTC(),
		account.ID,
		account.UserID,
	)
	if err != nil {
		return fmt.Errorf("repository: update account: %w", err)
	}
	return nil
}

// Delete elimina una cuenta por ID.
func (r *AccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM accounts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("repository: delete account: %w", err)
	}
	return nil
}

// scanner es una interfaz mínima que satisfacen pgx.Row y pgx.Rows.
type scanner interface {
	Scan(dest ...any) error
}

// scanAccount lee una fila de accounts y retorna el entity.
func scanAccount(row scanner) (entity.Account, error) {
	var (
		id           uuid.UUID
		userID       uuid.UUID
		name         string
		currencyCode string
		balanceStr   string
		createdAt    time.Time
		updatedAt    time.Time
	)

	if err := row.Scan(&id, &userID, &name, &currencyCode, &balanceStr, &createdAt, &updatedAt); err != nil {
		return entity.Account{}, fmt.Errorf("repository: scan account: %w", err)
	}

	amount, err := decimal.NewFromString(balanceStr)
	if err != nil {
		return entity.Account{}, fmt.Errorf("repository: parse balance %q: %w", balanceStr, err)
	}

	balance, err := valueobject.NewMoney(amount, currencyCode)
	if err != nil {
		return entity.Account{}, fmt.Errorf("repository: build money: %w", err)
	}

	return entity.Account{
		ID:           id,
		UserID:       userID,
		Name:         name,
		CurrencyCode: currencyCode,
		Balance:      balance,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, nil
}
