-- +goose Up
-- ENUM de períodos de presupuesto.
CREATE TYPE budget_period AS ENUM ('weekly', 'monthly', 'yearly');

-- Presupuestos por categoría y período.
-- limit_amount > 0: monto máximo permitido en el período.
-- CHECK (ends_at > starts_at): invariante de integridad — fin siempre posterior al inicio.
-- ON DELETE CASCADE en category_id: si se elimina la categoría, el budget pierde sentido.
CREATE TABLE budgets (
  id            UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id       UUID          NOT NULL,
  category_id   UUID          NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  currency_code CHAR(3)       NOT NULL REFERENCES currencies(code),
  limit_amount  NUMERIC(20,4) NOT NULL CHECK (limit_amount > 0),
  period        budget_period NOT NULL,
  starts_at     TIMESTAMPTZ   NOT NULL,
  ends_at       TIMESTAMPTZ   NOT NULL,
  created_at    TIMESTAMPTZ   NOT NULL DEFAULT now(),
  CHECK (ends_at > starts_at)
);
CREATE INDEX idx_budgets_user_id     ON budgets(user_id);
CREATE INDEX idx_budgets_category_id ON budgets(category_id);

-- +goose Down
DROP TABLE IF EXISTS budgets;
DROP TYPE  IF EXISTS budget_period;
