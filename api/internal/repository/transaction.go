package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/neto-app/neto/api/internal/domain/entity"
	domainrepo "github.com/neto-app/neto/api/internal/domain/repository"
	"github.com/neto-app/neto/api/internal/domain/valueobject"
)

// TransactionRepository implementa domain/repository.TransactionRepository usando pgx.
type TransactionRepository struct {
	pool *pgxpool.Pool
}

func newTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{pool: pool}
}

// Create inserta una transacción y actualiza el balance de la cuenta en la misma transacción DB.
func (r *TransactionRepository) Create(ctx context.Context, t entity.Transaction) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repository: begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	_, err = tx.Exec(ctx, `
		INSERT INTO transactions
			(id, user_id, account_id, category_id, amount, type, description, idempotency_key, occurred_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`,
		t.ID, t.UserID, t.AccountID, t.CategoryID,
		t.Amount.Amount.String(),
		string(t.Type), t.Description, t.IdempotencyKey,
		t.OccurredAt, t.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository: insert transaction: %w", err)
	}

	// Actualizar el balance de la cuenta en la misma transacción DB
	var balanceDelta string
	switch t.Type {
	case entity.Income:
		balanceDelta = "balance + " + t.Amount.Amount.String()
	case entity.Expense, entity.Transfer:
		balanceDelta = "balance - " + t.Amount.Amount.String()
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(`
		UPDATE accounts
		SET balance = %s, updated_at = $1
		WHERE id = $2 AND user_id = $3
	`, balanceDelta),
		time.Now().UTC(), t.AccountID, t.UserID,
	)
	if err != nil {
		return fmt.Errorf("repository: update account balance: %w", err)
	}

	return tx.Commit(ctx)
}

// GetByID busca una transacción por ID.
func (r *TransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.Transaction, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, account_id, category_id, amount, type, description, idempotency_key, occurred_at, created_at
		FROM transactions
		WHERE id = $1
	`, id)
	return scanTransaction(row)
}

// GetByUserID retorna transacciones del usuario aplicando filtros opcionales.
func (r *TransactionRepository) GetByUserID(ctx context.Context, userID uuid.UUID, filter domainrepo.TransactionFilter) ([]entity.Transaction, error) {
	args := []any{userID}
	conditions := []string{"user_id = $1"}
	idx := 2

	if filter.AccountID != nil {
		conditions = append(conditions, fmt.Sprintf("account_id = $%d", idx))
		args = append(args, *filter.AccountID)
		idx++
	}
	if filter.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", idx))
		args = append(args, *filter.CategoryID)
		idx++
	}
	if filter.From != nil {
		conditions = append(conditions, fmt.Sprintf("occurred_at >= $%d", idx))
		args = append(args, *filter.From)
		idx++
	}
	if filter.To != nil {
		conditions = append(conditions, fmt.Sprintf("occurred_at <= $%d", idx))
		args = append(args, *filter.To)
	}

	limit := 50
	if filter.Limit > 0 {
		limit = filter.Limit
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, account_id, category_id, amount, type, description, idempotency_key, occurred_at, created_at
		FROM transactions
		WHERE %s
		ORDER BY occurred_at DESC
		LIMIT %d OFFSET %d
	`, strings.Join(conditions, " AND "), limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("repository: list transactions: %w", err)
	}
	defer rows.Close()

	var transactions []entity.Transaction
	for rows.Next() {
		t, err := scanTransaction(rows)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}
	return transactions, rows.Err()
}

// GetByIdempotencyKey busca una transacción por su idempotency key.
func (r *TransactionRepository) GetByIdempotencyKey(ctx context.Context, key string) (entity.Transaction, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, account_id, category_id, amount, type, description, idempotency_key, occurred_at, created_at
		FROM transactions
		WHERE idempotency_key = $1
	`, key)
	return scanTransaction(row)
}

// Update actualiza los campos de una transacción.
func (r *TransactionRepository) Update(ctx context.Context, t entity.Transaction) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE transactions
		SET description = $1, category_id = $2
		WHERE id = $3 AND user_id = $4
	`, t.Description, t.CategoryID, t.ID, t.UserID)
	if err != nil {
		return fmt.Errorf("repository: update transaction: %w", err)
	}
	return nil
}

// Delete elimina una transacción por ID.
func (r *TransactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM transactions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("repository: delete transaction: %w", err)
	}
	return nil
}

// scanTransaction lee una fila de transactions.
func scanTransaction(row scanner) (entity.Transaction, error) {
	var (
		id             uuid.UUID
		userID         uuid.UUID
		accountID      uuid.UUID
		categoryID     *uuid.UUID
		amountStr      string
		txType         string
		description    string
		idempotencyKey string
		occurredAt     time.Time
		createdAt      time.Time
	)

	if err := row.Scan(
		&id, &userID, &accountID, &categoryID,
		&amountStr, &txType, &description, &idempotencyKey,
		&occurredAt, &createdAt,
	); err != nil {
		return entity.Transaction{}, fmt.Errorf("repository: scan transaction: %w", err)
	}

	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return entity.Transaction{}, fmt.Errorf("repository: parse transaction amount %q: %w", amountStr, err)
	}

	// Necesitamos la currency del account — usamos placeholder "unknown" aquí
	// porque el currency se obtiene del account. En queries reales hacer JOIN.
	// NOTA: Para producción, hacer JOIN con accounts para obtener currency_code.
	money, err := valueobject.NewMoney(amount, "ARS")
	if err != nil {
		return entity.Transaction{}, fmt.Errorf("repository: build transaction money: %w", err)
	}

	return entity.Transaction{
		ID:             id,
		UserID:         userID,
		AccountID:      accountID,
		CategoryID:     categoryID,
		Amount:         money,
		Type:           entity.TransactionType(txType),
		Description:    description,
		IdempotencyKey: idempotencyKey,
		OccurredAt:     occurredAt,
		CreatedAt:      createdAt,
	}, nil
}
