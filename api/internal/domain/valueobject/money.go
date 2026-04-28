// Package valueobject contiene los value objects del dominio de Neto.
// Los value objects son inmutables y se comparan por valor, no por identidad.
package valueobject

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// Money representa una cantidad de dinero en una moneda específica.
// Se usa decimal.Decimal internamente para evitar errores de punto flotante.
type Money struct {
	Amount       decimal.Decimal
	CurrencyCode string
}

// NewMoney crea un nuevo Money validando que la moneda no esté vacía y el monto sea >= 0.
func NewMoney(amount decimal.Decimal, currency string) (Money, error) {
	if currency == "" {
		return Money{}, fmt.Errorf("currency code cannot be empty")
	}
	if amount.IsNegative() {
		return Money{}, fmt.Errorf("amount cannot be negative: %s", amount.String())
	}
	return Money{Amount: amount, CurrencyCode: currency}, nil
}

// Add suma dos Money del mismo tipo de moneda.
func (m Money) Add(other Money) (Money, error) {
	if m.CurrencyCode != other.CurrencyCode {
		return Money{}, fmt.Errorf("currency mismatch: cannot add %s and %s", m.CurrencyCode, other.CurrencyCode)
	}
	return Money{Amount: m.Amount.Add(other.Amount), CurrencyCode: m.CurrencyCode}, nil
}

// Subtract resta two Money del mismo tipo de moneda. Retorna error si el resultado es negativo.
func (m Money) Subtract(other Money) (Money, error) {
	if m.CurrencyCode != other.CurrencyCode {
		return Money{}, fmt.Errorf("currency mismatch: cannot subtract %s from %s", other.CurrencyCode, m.CurrencyCode)
	}
	result := m.Amount.Sub(other.Amount)
	if result.IsNegative() {
		return Money{}, fmt.Errorf("subtraction result is negative: %s - %s", m.Amount.String(), other.Amount.String())
	}
	return Money{Amount: result, CurrencyCode: m.CurrencyCode}, nil
}

// IsZero retorna true si el monto es cero.
func (m Money) IsZero() bool {
	return m.Amount.IsZero()
}

// String retorna la representación legible del Money.
func (m Money) String() string {
	return fmt.Sprintf("%s %s", m.Amount.String(), m.CurrencyCode)
}
