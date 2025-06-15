-- Переименование столбца remaining_amount в remaining_debt в таблице credits
ALTER TABLE credits RENAME COLUMN remaining_amount TO remaining_debt;
