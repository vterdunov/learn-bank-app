-- Удаление добавленных полей из таблицы payment_schedules
ALTER TABLE payment_schedules
DROP COLUMN remaining_balance,
DROP COLUMN paid_date;
