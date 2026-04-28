-- +goose Up
-- Metas de ahorro del usuario.
-- current_amount >= 0: nunca negativo; target_amount > 0: siempre hay un objetivo.
-- deadline opcional — metas sin fecha límite son válidas.
-- El progreso (current/target * 100) se calcula en capa de aplicación, no se persiste.
CREATE TABLE goals (
  id             UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id        UUID          NOT NULL,
  name           VARCHAR(100)  NOT NULL,
  currency_code  CHAR(3)       NOT NULL REFERENCES currencies(code),
  target_amount  NUMERIC(20,4) NOT NULL CHECK (target_amount > 0),
  current_amount NUMERIC(20,4) NOT NULL DEFAULT 0 CHECK (current_amount >= 0),
  deadline       TIMESTAMPTZ,
  created_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ   NOT NULL DEFAULT now()
);
CREATE INDEX idx_goals_user_id ON goals(user_id);

-- +goose Down
DROP TABLE IF EXISTS goals;
