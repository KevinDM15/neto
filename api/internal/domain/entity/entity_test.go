package entity_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/neto-app/neto/api/internal/domain/entity"
	"github.com/neto-app/neto/api/internal/domain/valueobject"
)

// ---- Account ----

func TestNewAccount_Valid(t *testing.T) {
	userID := uuid.New()
	acc, err := entity.NewAccount(userID, "Checking", "USD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.Name != "Checking" {
		t.Errorf("expected name Checking, got %s", acc.Name)
	}
	if acc.Balance.CurrencyCode != "USD" {
		t.Errorf("expected USD balance, got %s", acc.Balance.CurrencyCode)
	}
	if !acc.Balance.IsZero() {
		t.Error("expected zero initial balance")
	}
}

func TestNewAccount_EmptyName(t *testing.T) {
	_, err := entity.NewAccount(uuid.New(), "", "USD")
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

// ---- Transaction ----

func TestNewTransaction_Valid(t *testing.T) {
	amount, _ := valueobject.NewMoney(decimal.NewFromFloat(50), "USD")
	tx, err := entity.NewTransaction(uuid.New(), uuid.New(), amount, entity.Income, "salary", "key-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.Type != entity.Income {
		t.Errorf("expected Income, got %s", tx.Type)
	}
}

func TestNewTransaction_InvalidType(t *testing.T) {
	amount, _ := valueobject.NewMoney(decimal.NewFromFloat(50), "USD")
	_, err := entity.NewTransaction(uuid.New(), uuid.New(), amount, "bad_type", "desc", "key-001")
	if err == nil {
		t.Fatal("expected error for invalid type, got nil")
	}
}

func TestNewTransaction_EmptyIdempotencyKey(t *testing.T) {
	amount, _ := valueobject.NewMoney(decimal.NewFromFloat(50), "USD")
	_, err := entity.NewTransaction(uuid.New(), uuid.New(), amount, entity.Expense, "desc", "")
	if err == nil {
		t.Fatal("expected error for empty idempotency key, got nil")
	}
}

// ---- Budget ----

func TestBudget_RemainingAmount(t *testing.T) {
	limit, _ := valueobject.NewMoney(decimal.NewFromFloat(1000), "USD")
	spent, _ := valueobject.NewMoney(decimal.NewFromFloat(300), "USD")

	b := entity.Budget{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		CategoryID: uuid.New(),
		Limit:      limit,
		Period:     entity.Monthly,
		StartsAt:   time.Now(),
		EndsAt:     time.Now().Add(30 * 24 * time.Hour),
	}

	remaining, err := b.RemainingAmount(spent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !remaining.Amount.Equal(decimal.NewFromFloat(700)) {
		t.Errorf("expected 700, got %s", remaining.Amount.String())
	}
}

func TestBudget_IsExceeded_True(t *testing.T) {
	limit, _ := valueobject.NewMoney(decimal.NewFromFloat(500), "USD")
	spent, _ := valueobject.NewMoney(decimal.NewFromFloat(600), "USD")
	b := entity.Budget{Limit: limit}
	exceeded, err := b.IsExceeded(spent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exceeded {
		t.Error("expected budget to be exceeded")
	}
}

func TestBudget_IsExceeded_False(t *testing.T) {
	limit, _ := valueobject.NewMoney(decimal.NewFromFloat(500), "USD")
	spent, _ := valueobject.NewMoney(decimal.NewFromFloat(200), "USD")
	b := entity.Budget{Limit: limit}
	exceeded, err := b.IsExceeded(spent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exceeded {
		t.Error("expected budget to not be exceeded")
	}
}

// ---- Debt ----

func TestDebt_IsPaid(t *testing.T) {
	now := time.Now().UTC()
	d := entity.Debt{PaidAt: &now}
	if !d.IsPaid() {
		t.Error("expected IsPaid to be true")
	}
}

func TestDebt_IsPaid_False(t *testing.T) {
	d := entity.Debt{}
	if d.IsPaid() {
		t.Error("expected IsPaid to be false")
	}
}

func TestDebt_IsOverdue_True(t *testing.T) {
	past := time.Now().Add(-48 * time.Hour)
	d := entity.Debt{DueDate: &past}
	if !d.IsOverdue() {
		t.Error("expected IsOverdue to be true")
	}
}

func TestDebt_IsOverdue_PaidDebt(t *testing.T) {
	past := time.Now().Add(-48 * time.Hour)
	now := time.Now()
	d := entity.Debt{DueDate: &past, PaidAt: &now}
	if d.IsOverdue() {
		t.Error("paid debt should not be overdue")
	}
}

// ---- Goal ----

func TestGoal_Progress_Zero(t *testing.T) {
	target, _ := valueobject.NewMoney(decimal.NewFromFloat(1000), "USD")
	g, _ := entity.NewGoal(uuid.New(), "Vacation", target)
	if g.Progress() != 0 {
		t.Errorf("expected 0%%, got %.2f%%", g.Progress())
	}
}

func TestGoal_Progress_Fifty(t *testing.T) {
	target, _ := valueobject.NewMoney(decimal.NewFromFloat(1000), "USD")
	current, _ := valueobject.NewMoney(decimal.NewFromFloat(500), "USD")
	g, _ := entity.NewGoal(uuid.New(), "Vacation", target)
	g.CurrentAmount = current
	if g.Progress() != 50 {
		t.Errorf("expected 50%%, got %.2f%%", g.Progress())
	}
}

func TestGoal_Progress_Completed(t *testing.T) {
	target, _ := valueobject.NewMoney(decimal.NewFromFloat(1000), "USD")
	current, _ := valueobject.NewMoney(decimal.NewFromFloat(1000), "USD")
	g, _ := entity.NewGoal(uuid.New(), "Vacation", target)
	g.CurrentAmount = current
	if g.Progress() != 100 {
		t.Errorf("expected 100%%, got %.2f%%", g.Progress())
	}
	if !g.IsCompleted() {
		t.Error("expected IsCompleted to be true")
	}
}
