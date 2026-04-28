-- +goose Up
-- Catálogo de monedas: referenciado por accounts, transactions, budgets, debts, goals.
-- code es PK natural (ISO 4217) — evita JOINs innecesarios en queries frecuentes.
CREATE TABLE currencies (
  code       CHAR(3)      PRIMARY KEY,
  name       VARCHAR(100) NOT NULL,
  symbol     VARCHAR(10)  NOT NULL,
  is_active  BOOLEAN      NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS currencies;
