-- Добавление недостающих полей в таблицу payment_schedules
ALTER TABLE payment_schedules
ADD COLUMN remaining_balance DECIMAL(15,2) NOT NULL DEFAULT 0.00,
ADD COLUMN paid_date TIMESTAMP NULL;

-- Добавление ограничений
ALTER TABLE payment_schedules
ADD CONSTRAINT chk_remaining_balance_non_negative CHECK (remaining_balance >= 0);

-- Комментарии к полям
COMMENT ON COLUMN payment_schedules.remaining_balance IS 'Остаток основного долга после платежа';
COMMENT ON COLUMN payment_schedules.paid_date IS 'Дата фактической оплаты (дубликат paid_at для совместимости)';
