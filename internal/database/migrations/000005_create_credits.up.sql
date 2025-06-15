-- Создание таблицы кредитов
CREATE TABLE IF NOT EXISTS credits (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    amount DECIMAL(15,2) NOT NULL,
    interest_rate DECIMAL(5,2) NOT NULL, -- Процентная ставка
    term_months INTEGER NOT NULL, -- Срок в месяцах
    monthly_payment DECIMAL(15,2) NOT NULL, -- Ежемесячный платеж (аннуитет)
    remaining_amount DECIMAL(15,2) NOT NULL, -- Остаток долга
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    start_date DATE NOT NULL DEFAULT CURRENT_DATE,
    end_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_credit_amount_positive CHECK (amount > 0),
    CONSTRAINT chk_interest_rate_valid CHECK (interest_rate >= 0 AND interest_rate <= 100),
    CONSTRAINT chk_term_positive CHECK (term_months > 0),
    CONSTRAINT chk_monthly_payment_positive CHECK (monthly_payment > 0),
    CONSTRAINT chk_remaining_amount_non_negative CHECK (remaining_amount >= 0),
    CONSTRAINT chk_credit_status_valid CHECK (
        status IN ('active', 'completed', 'overdue', 'cancelled')
    ),
    CONSTRAINT chk_end_date_after_start CHECK (end_date > start_date)
);

-- Создание индексов
CREATE INDEX IF NOT EXISTS idx_credits_user_id ON credits(user_id);
CREATE INDEX IF NOT EXISTS idx_credits_account_id ON credits(account_id);
CREATE INDEX IF NOT EXISTS idx_credits_status ON credits(status);
CREATE INDEX IF NOT EXISTS idx_credits_start_date ON credits(start_date);
CREATE INDEX IF NOT EXISTS idx_credits_end_date ON credits(end_date);

-- Триггер для автоматического обновления updated_at
CREATE TRIGGER update_credits_updated_at
    BEFORE UPDATE ON credits
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
