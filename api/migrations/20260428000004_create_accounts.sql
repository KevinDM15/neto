-- +goose Up
-- Cuentas bancarias/wallet del usuario.
-- balance se mantiene actualizado vía triggers o capa de aplicación al procesar transacciones.
-- NUMERIC(20,4) — 4 decimales suficientes para cualquier moneda LATAM + USD/EUR.
CREATE TABLE accounts (
  id            UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id       UUID          NOT NULL,
  name          VARCHAR(100)  NOT NULL,
  currency_code CHAR(3)       NOT NULL REFERENCES currencies(code),
  balance       NUMERIC(20,4) NOT NULL DEFAULT 0,
  created_at    TIMESTAMPTZ   NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ   NOT NULL DEFAULT now()
);
CREATE INDEX idx_accounts_user_id ON accounts(user_id);

-- +goose Down
DROP TABLE IF EXISTS accounts;
