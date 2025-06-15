-- Удаление таблицы банковских карт
DROP TRIGGER IF EXISTS update_cards_updated_at ON cards;
DROP TABLE IF EXISTS cards CASCADE;
