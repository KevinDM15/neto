-- +goose Up
-- Cache de idempotencia para endpoints mutantes (POST).
-- key es el UUID v4 provisto por el cliente en el header Idempotency-Key.
-- response_body almacena el JSON serializado para replay exacto.
-- expires_at DEFAULT now() + 24h: ventana estándar de idempotencia para APIs REST.
-- idx_expires_at: usado por el job de limpieza periódica (DELETE WHERE expires_at < now()).
CREATE TABLE idempotency_keys (
  key              VARCHAR(36)  PRIMARY KEY,
  user_id          UUID         NOT NULL,
  request_path     VARCHAR(255) NOT NULL,
  response_status  INT          NOT NULL,
  response_body    TEXT         NOT NULL,
  created_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
  expires_at       TIMESTAMPTZ  NOT NULL DEFAULT now() + INTERVAL '24 hours'
);
CREATE INDEX idx_idempotency_keys_expires_at ON idempotency_keys(expires_at);

-- +goose Down
DROP TABLE IF EXISTS idempotency_keys;
