package valueobject

import (
	"fmt"
	"strings"
)

// Currency representa una moneda soportada por el sistema.
type Currency struct {
	Code     string
	Name     string
	Symbol   string
	IsActive bool
}

// NewCurrency crea una nueva Currency validando que el código tenga exactamente 3 caracteres en mayúsculas.
func NewCurrency(code, name, symbol string) (Currency, error) {
	if len(code) != 3 {
		return Currency{}, fmt.Errorf("currency code must be exactly 3 characters, got %d: %q", len(code), code)
	}
	if code != strings.ToUpper(code) {
		return Currency{}, fmt.Errorf("currency code must be uppercase, got: %q", code)
	}
	if name == "" {
		return Currency{}, fmt.Errorf("currency name cannot be empty")
	}
	return Currency{Code: code, Name: name, Symbol: symbol, IsActive: true}, nil
}
