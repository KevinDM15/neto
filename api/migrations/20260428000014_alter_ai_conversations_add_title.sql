-- +goose Up
ALTER TABLE ai_conversations
  ADD COLUMN title      TEXT,
  ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- +goose Down
ALTER TABLE ai_conversations
  DROP COLUMN IF EXISTS title,
  DROP COLUMN IF EXISTS updated_at;
