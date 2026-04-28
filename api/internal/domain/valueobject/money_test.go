package valueobject_test

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/neto-app/neto/api/internal/domain/valueobject"
)

func TestNewMoney_Valid(t *testing.T) {
	m, err := valueobject.NewMoney(decimal.NewFromFloat(100.50), "USD")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if m.CurrencyCode != "USD" {
		t.Errorf("expected USD, got %s", m.CurrencyCode)
	}
}

func TestNewMoney_NegativeAmount(t *testing.T) {
	_, err := valueobject.NewMoney(decimal.NewFromFloat(-1), "USD")
	if err == nil {
		t.Fatal("expected error for negative amount, got nil")
	}
}

func TestNewMoney_EmptyCurrency(t *testing.T) {
	_, err := valueobject.NewMoney(decimal.NewFromFloat(10), "")
	if err == nil {
		t.Fatal("expected error for empty currency, got nil")
	}
}

func TestMoney_Add_SameCurrency(t *testing.T) {
	a, _ := valueobject.NewMoney(decimal.NewFromFloat(10), "USD")
	b, _ := valueobject.NewMoney(decimal.NewFromFloat(5), "USD")
	result, err := a.Add(b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Amount.Equal(decimal.NewFromFloat(15)) {
		t.Errorf("expected 15, got %s", result.Amount.String())
	}
}

func TestMoney_Add_DifferentCurrency(t *testing.T) {
	a, _ := valueobject.NewMoney(decimal.NewFromFloat(10), "USD")
	b, _ := valueobject.NewMoney(decimal.NewFromFloat(5), "ARS")
	_, err := a.Add(b)
	if err == nil {
		t.Fatal("expected error when adding different currencies, got nil")
	}
}

func TestMoney_Subtract_SameCurrency(t *testing.T) {
	a, _ := valueobject.NewMoney(decimal.NewFromFloat(10), "USD")
	b, _ := valueobject.NewMoney(decimal.NewFromFloat(3), "USD")
	result, err := a.Subtract(b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Amount.Equal(decimal.NewFromFloat(7)) {
		t.Errorf("expected 7, got %s", result.Amount.String())
	}
}

func TestMoney_Subtract_DifferentCurrency(t *testing.T) {
	a, _ := valueobject.NewMoney(decimal.NewFromFloat(10), "USD")
	b, _ := valueobject.NewMoney(decimal.NewFromFloat(3), "EUR")
	_, err := a.Subtract(b)
	if err == nil {
		t.Fatal("expected error when subtracting different currencies, got nil")
	}
}

func TestMoney_Subtract_ResultNegative(t *testing.T) {
	a, _ := valueobject.NewMoney(decimal.NewFromFloat(3), "USD")
	b, _ := valueobject.NewMoney(decimal.NewFromFloat(10), "USD")
	_, err := a.Subtract(b)
	if err == nil {
		t.Fatal("expected error when result is negative, got nil")
	}
}

func TestMoney_IsZero(t *testing.T) {
	m, _ := valueobject.NewMoney(decimal.Zero, "USD")
	if !m.IsZero() {
		t.Error("expected IsZero to be true")
	}
}

func TestNewCurrency_Valid(t *testing.T) {
	c, err := valueobject.NewCurrency("USD", "US Dollar", "$")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if c.Code != "USD" {
		t.Errorf("expected USD, got %s", c.Code)
	}
	if !c.IsActive {
		t.Error("expected IsActive to be true")
	}
}

func TestNewCurrency_InvalidCode_TooShort(t *testing.T) {
	_, err := valueobject.NewCurrency("US", "US Dollar", "$")
	if err == nil {
		t.Fatal("expected error for 2-char code, got nil")
	}
}

func TestNewCurrency_InvalidCode_TooLong(t *testing.T) {
	_, err := valueobject.NewCurrency("USDD", "US Dollar", "$")
	if err == nil {
		t.Fatal("expected error for 4-char code, got nil")
	}
}

func TestNewCurrency_InvalidCode_Lowercase(t *testing.T) {
	_, err := valueobject.NewCurrency("usd", "US Dollar", "$")
	if err == nil {
		t.Fatal("expected error for lowercase code, got nil")
	}
}
