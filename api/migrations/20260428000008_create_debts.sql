-- +goose Up
-- ENUM de dirección de deuda desde la perspectiva del usuario.
-- 'owed': alguien me debe a mí. 'owing': yo le debo a alguien.
CREATE TYPE debt_direction AS ENUM ('owed', 'owing');

-- Registro de deudas con terceros.
-- paid_at NULL = deuda activa; NOT NULL = saldada.
-- due_date opcional — no todas las deudas tienen vencimiento.
-- amount > 0: siempre positivo; la dirección define quién debe a quién.
CREATE TABLE debts (
  id                UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id           UUID           NOT NULL,
  counterparty_name VARCHAR(100)   NOT NULL,
  currency_code     CHAR(3)        NOT NULL REFERENCES currencies(code),
  amount            NUMERIC(20,4)  NOT NULL CHECK (amount > 0),
  direction         debt_direction NOT NULL,
  due_date          TIMESTAMPTZ,
  paid_at           TIMESTAMPTZ,
  notes             TEXT,
  created_at        TIMESTAMPTZ    NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ    NOT NULL DEFAULT now()
);
CREATE INDEX idx_debts_user_id ON debts(user_id);

-- +goose Down
DROP TABLE IF EXISTS debts;
DROP TYPE  IF EXISTS debt_direction;
