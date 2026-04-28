-- +goose Up
-- Tasas de cambio entre pares de monedas.
-- UNIQUE (from_code, to_code) — solo un rate vigente por par.
-- rate > 0 garantizado por CHECK; scale de 8 decimales para crypto-ready.
CREATE TABLE exchange_rates (
  id          UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  from_code   CHAR(3)       NOT NULL REFERENCES currencies(code),
  to_code     CHAR(3)       NOT NULL REFERENCES currencies(code),
  rate        NUMERIC(20,8) NOT NULL CHECK (rate > 0),
  fetched_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
  UNIQUE (from_code, to_code)
);
CREATE INDEX idx_exchange_rates_from ON exchange_rates(from_code);
CREATE INDEX idx_exchange_rates_to   ON exchange_rates(to_code);

-- +goose Down
DROP TABLE IF EXISTS exchange_rates;
