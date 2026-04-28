package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neto-app/neto/api/internal/domain/valueobject"
)

// CurrencyRepository implementa domain/repository.CurrencyRepository usando pgx.
type CurrencyRepository struct {
	pool *pgxpool.Pool
}

func newCurrencyRepository(pool *pgxpool.Pool) *CurrencyRepository {
	return &CurrencyRepository{pool: pool}
}

// GetAll retorna todas las monedas registradas.
func (r *CurrencyRepository) GetAll(ctx context.Context) ([]valueobject.Currency, error) {
	return r.query(ctx, `SELECT code, name, symbol, is_active FROM currencies ORDER BY code`)
}

// GetByCode busca una moneda por su código ISO 4217.
func (r *CurrencyRepository) GetByCode(ctx context.Context, code string) (valueobject.Currency, error) {
	row := r.pool.QueryRow(ctx, `SELECT code, name, symbol, is_active FROM currencies WHERE code = $1`, code)

	var c valueobject.Currency
	if err := row.Scan(&c.Code, &c.Name, &c.Symbol, &c.IsActive); err != nil {
		return valueobject.Currency{}, fmt.Errorf("repository: get currency by code %q: %w", code, err)
	}
	return c, nil
}

// GetActive retorna solo las monedas activas.
func (r *CurrencyRepository) GetActive(ctx context.Context) ([]valueobject.Currency, error) {
	return r.query(ctx, `SELECT code, name, symbol, is_active FROM currencies WHERE is_active = true ORDER BY code`)
}

// query ejecuta una query que retorna múltiples monedas.
func (r *CurrencyRepository) query(ctx context.Context, sql string, args ...any) ([]valueobject.Currency, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("repository: query currencies: %w", err)
	}
	defer rows.Close()

	var currencies []valueobject.Currency
	for rows.Next() {
		var c valueobject.Currency
		if err := rows.Scan(&c.Code, &c.Name, &c.Symbol, &c.IsActive); err != nil {
			return nil, fmt.Errorf("repository: scan currency: %w", err)
		}
		currencies = append(currencies, c)
	}
	return currencies, rows.Err()
}
