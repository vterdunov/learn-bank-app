-- Откат переименования столбца remaining_debt обратно в remaining_amount
ALTER TABLE credits RENAME COLUMN remaining_debt TO remaining_amount;
