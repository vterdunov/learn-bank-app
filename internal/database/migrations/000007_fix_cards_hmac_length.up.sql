-- Увеличиваем размер поля hmac в таблице cards
ALTER TABLE cards ALTER COLUMN hmac TYPE TEXT;
