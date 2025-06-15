-- Возвращаем размер поля hmac обратно
ALTER TABLE cards ALTER COLUMN hmac TYPE VARCHAR(64);
