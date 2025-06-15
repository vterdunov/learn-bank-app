-- Создание таблицы банковских карт
CREATE TABLE IF NOT EXISTS cards (
    id SERIAL PRIMARY KEY,
    account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    encrypted_data TEXT NOT NULL, -- Зашифрованные номер карты и срок действия
    hmac VARCHAR(64) NOT NULL, -- HMAC для проверки целостности
    cvv_hash VARCHAR(255) NOT NULL, -- Хеш CVV кода
    expiry_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_card_status_valid CHECK (status IN ('active', 'blocked', 'expired', 'cancelled'))
);

-- Создание индексов
CREATE INDEX IF NOT EXISTS idx_cards_account_id ON cards(account_id);
CREATE INDEX IF NOT EXISTS idx_cards_status ON cards(status);
CREATE INDEX IF NOT EXISTS idx_cards_expiry_date ON cards(expiry_date);

-- Триггер для автоматического обновления updated_at
CREATE TRIGGER update_cards_updated_at
    BEFORE UPDATE ON cards
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
