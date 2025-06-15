-- Создание таблицы графика платежей
CREATE TABLE IF NOT EXISTS payment_schedules (
    id SERIAL PRIMARY KEY,
    credit_id INTEGER NOT NULL REFERENCES credits(id) ON DELETE CASCADE,
    payment_number INTEGER NOT NULL, -- Номер платежа
    due_date DATE NOT NULL, -- Дата платежа
    payment_amount DECIMAL(15,2) NOT NULL, -- Сумма платежа
    principal_amount DECIMAL(15,2) NOT NULL, -- Основной долг
    interest_amount DECIMAL(15,2) NOT NULL, -- Проценты
    penalty_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00, -- Штраф за просрочку
    paid_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00, -- Фактически выплачено
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    paid_at TIMESTAMP NULL, -- Дата фактической оплаты
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_payment_number_positive CHECK (payment_number > 0),
    CONSTRAINT chk_payment_amount_positive CHECK (payment_amount > 0),
    CONSTRAINT chk_principal_amount_positive CHECK (principal_amount > 0),
    CONSTRAINT chk_interest_amount_non_negative CHECK (interest_amount >= 0),
    CONSTRAINT chk_penalty_amount_non_negative CHECK (penalty_amount >= 0),
    CONSTRAINT chk_paid_amount_non_negative CHECK (paid_amount >= 0),
    CONSTRAINT chk_payment_status_valid CHECK (
        status IN ('pending', 'paid', 'overdue', 'partially_paid')
    ),
    CONSTRAINT unique_credit_payment_number UNIQUE (credit_id, payment_number)
);

-- Создание индексов
CREATE INDEX IF NOT EXISTS idx_payment_schedules_credit_id ON payment_schedules(credit_id);
CREATE INDEX IF NOT EXISTS idx_payment_schedules_due_date ON payment_schedules(due_date);
CREATE INDEX IF NOT EXISTS idx_payment_schedules_status ON payment_schedules(status);
CREATE INDEX IF NOT EXISTS idx_payment_schedules_paid_at ON payment_schedules(paid_at);

-- Композитный индекс для поиска просроченных платежей
CREATE INDEX IF NOT EXISTS idx_payment_schedules_overdue ON payment_schedules(
    status, due_date
) WHERE status IN ('pending', 'partially_paid');

-- Триггер для автоматического обновления updated_at
CREATE TRIGGER update_payment_schedules_updated_at
    BEFORE UPDATE ON payment_schedules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
