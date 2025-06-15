-- Удаление таблицы графика платежей
DROP TRIGGER IF EXISTS update_payment_schedules_updated_at ON payment_schedules;
DROP TABLE IF EXISTS payment_schedules CASCADE;
