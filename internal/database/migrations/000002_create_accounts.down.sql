-- Удаление таблицы банковских счетов
DROP TRIGGER IF EXISTS update_accounts_updated_at ON accounts;
DROP TABLE IF EXISTS accounts CASCADE;
