-- +goose Up
-- ENUM de roles en conversación IA — alineado con OpenAI/Anthropic API contracts.
CREATE TYPE message_role AS ENUM ('user', 'assistant');

-- Conversaciones con el asistente IA.
-- Separar conversation de messages permite múltiples turnos sin repetir metadata.
CREATE TABLE ai_conversations (
  id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id    UUID        NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Mensajes individuales dentro de una conversación.
-- ON DELETE CASCADE: si se borra la conversación, se borran todos sus mensajes.
-- content TEXT: sin límite fijo — respuestas largas del asistente son esperadas.
CREATE TABLE ai_messages (
  id              UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
  conversation_id UUID         NOT NULL REFERENCES ai_conversations(id) ON DELETE CASCADE,
  role            message_role NOT NULL,
  content         TEXT         NOT NULL,
  created_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE INDEX idx_ai_conversations_user_id       ON ai_conversations(user_id);
CREATE INDEX idx_ai_messages_conversation_id    ON ai_messages(conversation_id);

-- +goose Down
DROP TABLE IF EXISTS ai_messages;
DROP TABLE IF EXISTS ai_conversations;
DROP TYPE  IF EXISTS message_role;
