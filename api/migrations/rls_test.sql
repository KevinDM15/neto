-- =============================================================================
-- rls_test.sql — Verificación manual de aislamiento RLS
-- =============================================================================
-- Cómo usar:
--   1. Abrí Supabase Dashboard → SQL Editor
--   2. Ejecutá este script COMPLETO de una sola vez
--   3. Revisá los resultados: todas las queries de verificación deben retornar 0 filas
--
-- IMPORTANTE: Este script simula dos usuarios con UUIDs fijos.
--   user_a_id: el usuario "atacante" que intenta leer datos ajenos
--   user_b_id: el usuario "víctima" cuyos datos deben estar protegidos
--
-- El script usa set_config para simular auth.uid() de Supabase.
-- En producción, auth.uid() viene del JWT decodificado por Supabase Auth.
-- =============================================================================

-- Usuarios de prueba
DO $$
DECLARE
  user_a_id UUID := '00000000-0000-0000-0000-000000000001';
  user_b_id UUID := '00000000-0000-0000-0000-000000000002';
  dummy_currency CHAR(3) := 'USD';
  account_b_id UUID;
  category_b_id UUID;
  conv_b_id UUID;
BEGIN

  -- Asegurarse que existe la moneda de prueba
  INSERT INTO currencies (code, name, symbol)
  VALUES ('USD', 'US Dollar', '$')
  ON CONFLICT DO NOTHING;

  -- =========================================================================
  -- SETUP: Crear datos para user_b (la víctima)
  -- =========================================================================

  -- Simular que somos user_b para insertar datos
  PERFORM set_config('request.jwt.claims', json_build_object('sub', user_b_id)::text, true);

  -- Cuenta de user_b
  INSERT INTO accounts (user_id, name, currency_code, balance)
  VALUES (user_b_id, 'Cuenta B', dummy_currency, 1000)
  RETURNING id INTO account_b_id;

  -- Categoría de user_b
  INSERT INTO categories (user_id, name)
  VALUES (user_b_id, 'Categoría B')
  RETURNING id INTO category_b_id;

  -- Transacción de user_b
  INSERT INTO transactions (user_id, account_id, amount, currency_code, type, idempotency_key)
  VALUES (user_b_id, account_b_id, 500, dummy_currency, 'income', 'test-idempotency-key-b-001');

  -- Budget de user_b
  INSERT INTO budgets (user_id, category_id, currency_code, limit_amount, period, starts_at, ends_at)
  VALUES (user_b_id, category_b_id, dummy_currency, 300, 'monthly', now(), now() + INTERVAL '1 month');

  -- Deuda de user_b
  INSERT INTO debts (user_id, counterparty_name, currency_code, amount, direction)
  VALUES (user_b_id, 'Juan', dummy_currency, 100, 'owed');

  -- Meta de user_b
  INSERT INTO goals (user_id, name, currency_code, target_amount)
  VALUES (user_b_id, 'Meta B', dummy_currency, 5000);

  -- Conversación IA de user_b
  INSERT INTO ai_conversations (user_id)
  VALUES (user_b_id)
  RETURNING id INTO conv_b_id;

  INSERT INTO ai_messages (conversation_id, role, content)
  VALUES (conv_b_id, 'user', 'Mensaje privado de user_b');

  -- Idempotency key de user_b
  INSERT INTO idempotency_keys (key, user_id, request_path, response_status, response_body)
  VALUES ('rls-test-key-b-001', user_b_id, '/api/v1/transactions', 201, '{"id":"..."}');

  RAISE NOTICE '✅ Setup: datos de user_b creados correctamente';

  -- =========================================================================
  -- ATAQUE: Simular que somos user_a intentando leer datos de user_b
  -- =========================================================================

  PERFORM set_config('request.jwt.claims', json_build_object('sub', user_a_id)::text, true);

  -- CHECK 1: accounts — user_a no debe ver cuentas de user_b
  IF (SELECT COUNT(*) FROM accounts WHERE user_id = user_b_id) > 0 THEN
    RAISE WARNING '❌ FALLO: accounts — user_a puede ver cuentas de user_b';
  ELSE
    RAISE NOTICE '✅ accounts: aislamiento OK';
  END IF;

  -- CHECK 2: categories — user_a no debe ver categorías de user_b
  IF (SELECT COUNT(*) FROM categories WHERE user_id = user_b_id) > 0 THEN
    RAISE WARNING '❌ FALLO: categories — user_a puede ver categorías de user_b';
  ELSE
    RAISE NOTICE '✅ categories: aislamiento OK';
  END IF;

  -- CHECK 3: transactions — user_a no debe ver transacciones de user_b
  IF (SELECT COUNT(*) FROM transactions WHERE user_id = user_b_id) > 0 THEN
    RAISE WARNING '❌ FALLO: transactions — user_a puede ver transacciones de user_b';
  ELSE
    RAISE NOTICE '✅ transactions: aislamiento OK';
  END IF;

  -- CHECK 4: budgets — user_a no debe ver presupuestos de user_b
  IF (SELECT COUNT(*) FROM budgets WHERE user_id = user_b_id) > 0 THEN
    RAISE WARNING '❌ FALLO: budgets — user_a puede ver presupuestos de user_b';
  ELSE
    RAISE NOTICE '✅ budgets: aislamiento OK';
  END IF;

  -- CHECK 5: debts — user_a no debe ver deudas de user_b
  IF (SELECT COUNT(*) FROM debts WHERE user_id = user_b_id) > 0 THEN
    RAISE WARNING '❌ FALLO: debts — user_a puede ver deudas de user_b';
  ELSE
    RAISE NOTICE '✅ debts: aislamiento OK';
  END IF;

  -- CHECK 6: goals — user_a no debe ver metas de user_b
  IF (SELECT COUNT(*) FROM goals WHERE user_id = user_b_id) > 0 THEN
    RAISE WARNING '❌ FALLO: goals — user_a puede ver metas de user_b';
  ELSE
    RAISE NOTICE '✅ goals: aislamiento OK';
  END IF;

  -- CHECK 7: ai_conversations — user_a no debe ver conversaciones de user_b
  IF (SELECT COUNT(*) FROM ai_conversations WHERE user_id = user_b_id) > 0 THEN
    RAISE WARNING '❌ FALLO: ai_conversations — user_a puede ver conversaciones de user_b';
  ELSE
    RAISE NOTICE '✅ ai_conversations: aislamiento OK';
  END IF;

  -- CHECK 8: ai_messages — user_a no debe ver mensajes de conversaciones de user_b
  IF (SELECT COUNT(*) FROM ai_messages WHERE conversation_id = conv_b_id) > 0 THEN
    RAISE WARNING '❌ FALLO: ai_messages — user_a puede ver mensajes de user_b';
  ELSE
    RAISE NOTICE '✅ ai_messages: aislamiento OK (via JOIN a ai_conversations)';
  END IF;

  -- CHECK 9: idempotency_keys — user_a no debe ver keys de user_b
  IF (SELECT COUNT(*) FROM idempotency_keys WHERE user_id = user_b_id) > 0 THEN
    RAISE WARNING '❌ FALLO: idempotency_keys — user_a puede ver keys de user_b';
  ELSE
    RAISE NOTICE '✅ idempotency_keys: aislamiento OK';
  END IF;

  -- =========================================================================
  -- CLEANUP: Eliminar datos de prueba (usando service_role implícito en SQL Editor)
  -- =========================================================================
  -- Nota: Para limpiar, es posible que necesités deshabilitar RLS temporalmente
  -- o conectarte como service_role. Los datos de prueba tienen UUIDs fijos y
  -- pueden eliminarse manualmente si es necesario.

  RAISE NOTICE '✅ Verificación RLS completa. Revisá los mensajes NOTICE/WARNING arriba.';

END $$;
