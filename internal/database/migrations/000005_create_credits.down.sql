-- Удаление таблицы кредитов
DROP TRIGGER IF EXISTS update_credits_updated_at ON credits;
DROP TABLE IF EXISTS credits CASCADE;
