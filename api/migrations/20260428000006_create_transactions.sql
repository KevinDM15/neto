-- +goose Up
-- ENUM de tipos de transacción — extensible si se agregan tipos futuros via ALTER TYPE.
CREATE TYPE transaction_type AS ENUM ('income', 'expense', 'transfer');

-- Transacciones financieras del usuario.
-- idempotency_key: UUID v4 generado por el cliente para evitar duplicados en reintentos.
-- UNIQUE (user_id, idempotency_key) — scope por usuario para no colisionar entre users.
-- amount > 0: el signo no va en amount, sino en type (income suma, expense resta).
-- ON DELETE CASCADE en account_id: si se borra la cuenta, se borran sus transacciones.
-- ON DELETE SET NULL en category_id: categoría opcional; si se borra, queda sin categoría.
-- idx_transactions_occurred_at DESC: queries de historial siempre ordenan por fecha desc.
CREATE TABLE transactions (
  id               UUID             PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id          UUID             NOT NULL,
  account_id       UUID             NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  category_id      UUID             REFERENCES categories(id) ON DELETE SET NULL,
  amount           NUMERIC(20,4)    NOT NULL CHECK (amount > 0),
  currency_code    CHAR(3)          NOT NULL REFERENCES currencies(code),
  type             transaction_type NOT NULL,
  description      TEXT,
  idempotency_key  VARCHAR(36)      NOT NULL,
  occurred_at      TIMESTAMPTZ      NOT NULL DEFAULT now(),
  created_at       TIMESTAMPTZ      NOT NULL DEFAULT now(),
  UNIQUE (user_id, idempotency_key)
);
CREATE INDEX idx_transactions_user_id     ON transactions(user_id);
CREATE INDEX idx_transactions_account_id  ON transactions(account_id);
CREATE INDEX idx_transactions_occurred_at ON transactions(occurred_at DESC);

-- +goose Down
DROP TABLE IF EXISTS transactions;
DROP TYPE  IF EXISTS transaction_type;
