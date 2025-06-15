-- Создание таблицы транзакций
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    from_account_id INTEGER REFERENCES accounts(id) ON DELETE SET NULL,
    to_account_id INTEGER REFERENCES accounts(id) ON DELETE SET NULL,
    amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'RUB',
    transaction_type VARCHAR(20) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_amount_positive CHECK (amount > 0),
    CONSTRAINT chk_currency_valid CHECK (currency IN ('RUB')),
    CONSTRAINT chk_transaction_type_valid CHECK (
        transaction_type IN ('deposit', 'withdrawal', 'transfer', 'payment', 'credit_payment', 'penalty')
    ),
    CONSTRAINT chk_status_valid CHECK (status IN ('pending', 'completed', 'failed', 'cancelled')),
    CONSTRAINT chk_accounts_not_same CHECK (
        (from_account_id IS NULL) OR
        (to_account_id IS NULL) OR
        (from_account_id != to_account_id)
    )
);

-- Создание индексов
CREATE INDEX IF NOT EXISTS idx_transactions_from_account ON transactions(from_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_to_account ON transactions(to_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions(transaction_type);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at);

-- Композитный индекс для аналитики по счету
CREATE INDEX IF NOT EXISTS idx_transactions_account_date ON transactions(
    from_account_id, created_at
) WHERE from_account_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_transactions_account_to_date ON transactions(
    to_account_id, created_at
) WHERE to_account_id IS NOT NULL;
