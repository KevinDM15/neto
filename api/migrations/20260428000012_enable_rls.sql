-- +goose Up

-- =============================================================================
-- Row Level Security (RLS) — Aislamiento por usuario
-- =============================================================================
-- Todas las tablas con datos de usuario tienen RLS habilitado.
-- Las políticas usan auth.uid() de Supabase, que retorna el UUID del JWT.
-- El service_role bypasea RLS — usarlo solo en workers/jobs, nunca en la API pública.
-- =============================================================================

-- Habilitar RLS en todas las tablas de usuario
ALTER TABLE accounts         ENABLE ROW LEVEL SECURITY;
ALTER TABLE categories       ENABLE ROW LEVEL SECURITY;
ALTER TABLE transactions     ENABLE ROW LEVEL SECURITY;
ALTER TABLE budgets          ENABLE ROW LEVEL SECURITY;
ALTER TABLE debts            ENABLE ROW LEVEL SECURITY;
ALTER TABLE goals            ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_conversations ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_messages      ENABLE ROW LEVEL SECURITY;
ALTER TABLE idempotency_keys ENABLE ROW LEVEL SECURITY;

-- accounts: solo el dueño ve sus cuentas
CREATE POLICY accounts_user_isolation ON accounts
  USING (user_id = auth.uid());

-- categories: solo el dueño ve sus categorías
CREATE POLICY categories_user_isolation ON categories
  USING (user_id = auth.uid());

-- transactions: solo el dueño ve sus transacciones
CREATE POLICY transactions_user_isolation ON transactions
  USING (user_id = auth.uid());

-- budgets: solo el dueño ve sus presupuestos
CREATE POLICY budgets_user_isolation ON budgets
  USING (user_id = auth.uid());

-- debts: solo el dueño ve sus deudas
CREATE POLICY debts_user_isolation ON debts
  USING (user_id = auth.uid());

-- goals: solo el dueño ve sus metas
CREATE POLICY goals_user_isolation ON goals
  USING (user_id = auth.uid());

-- ai_conversations: solo el dueño ve sus conversaciones
CREATE POLICY ai_conversations_user_isolation ON ai_conversations
  USING (user_id = auth.uid());

-- ai_messages: aislamiento indirecto via JOIN a ai_conversations.
-- No tiene user_id propio — hereda el aislamiento de su conversación padre.
CREATE POLICY ai_messages_user_isolation ON ai_messages
  USING (
    conversation_id IN (
      SELECT id FROM ai_conversations WHERE user_id = auth.uid()
    )
  );

-- idempotency_keys: cada usuario solo accede a sus propias keys
CREATE POLICY idempotency_keys_user_isolation ON idempotency_keys
  USING (user_id = auth.uid());

-- +goose Down
DROP POLICY IF EXISTS accounts_user_isolation         ON accounts;
DROP POLICY IF EXISTS categories_user_isolation       ON categories;
DROP POLICY IF EXISTS transactions_user_isolation     ON transactions;
DROP POLICY IF EXISTS budgets_user_isolation          ON budgets;
DROP POLICY IF EXISTS debts_user_isolation            ON debts;
DROP POLICY IF EXISTS goals_user_isolation            ON goals;
DROP POLICY IF EXISTS ai_conversations_user_isolation ON ai_conversations;
DROP POLICY IF EXISTS ai_messages_user_isolation      ON ai_messages;
DROP POLICY IF EXISTS idempotency_keys_user_isolation ON idempotency_keys;

ALTER TABLE accounts         DISABLE ROW LEVEL SECURITY;
ALTER TABLE categories       DISABLE ROW LEVEL SECURITY;
ALTER TABLE transactions     DISABLE ROW LEVEL SECURITY;
ALTER TABLE budgets          DISABLE ROW LEVEL SECURITY;
ALTER TABLE debts            DISABLE ROW LEVEL SECURITY;
ALTER TABLE goals            DISABLE ROW LEVEL SECURITY;
ALTER TABLE ai_conversations DISABLE ROW LEVEL SECURITY;
ALTER TABLE ai_messages      DISABLE ROW LEVEL SECURITY;
ALTER TABLE idempotency_keys DISABLE ROW LEVEL SECURITY;
