# Bank App

[![CI](https://github.com/vterdunov/learn-bank-app/actions/workflows/ci.yml/badge.svg)](https://github.com/vterdunov/learn-bank-app/actions/workflows/ci.yml)

REST API для банковского сервиса на Go с использованием чистой архитектуры.

## Содержание

- [Описание проекта](#описание-проекта)
- [Архитектура](#архитектура)
- [Быстрый запуск](#быстрый-запуск)
- [API Endpoints](#api-endpoints)
- [Тестирование](#тестирование)
- [Функциональные особенности](#функциональные-особенности)
- [Технологии и библиотеки](#технологии-и-библиотеки)
- [Структура базы данных](#структура-базы-данных)
- [Разработка](#разработка)
- [Чеклист выполненных работ](#чеклист-выполненных-работ)

## Описание проекта
- **Аутентификация и авторизация** с JWT токенами
- **Управление банковскими счетами** (создание, пополнение, списание)
- **Переводы между счетами** с поддержкой транзакций
- **Выпуск виртуальных карт** с шифрованием данных
- **Кредитные операции** с расчетом аннуитетных платежей
- **Аналитика и прогнозирование** финансовых операций
- **Email уведомления** о важных операциях
- **Интеграция с ЦБ РФ** для получения ключевой ставки
- **Высокий уровень безопасности** (шифрование, хеширование, HMAC)

## Архитектура

Приложение построено по принципам чистой архитектуры:

```
cmd/app/          - точка входа приложения
internal/
├── config/       - конфигурация приложения
├── domain/       - доменные модели и бизнес-логика
├── handlers/     - HTTP обработчики
├── middleware/   - промежуточные обработчики
├── repository/   - слой доступа к данным
├── router/       - маршрутизация
├── service/      - бизнес-логика и сервисы
└── utils/        - вспомогательные утилиты
pkg/logger/       - логирование
```

## Быстрый запуск

### Предварительные требования

- Go 1.24+
- PostgreSQL 17
- Docker

### Установка и запуск

1. **Клонируйте репозиторий:**
```bash
git clone https://github.com/vterdunov/learn-bank-app.git
cd learn-bank-app
```

2. **Запустите PostgreSQL через Docker:**
```bash
cd deployments
docker-compose up -d
```

3. **Скопируйте файл с переменными окружения:**
```bash
cp .env.example .env
```

4. **Отредактируйте `.env` файл:**
```bash
# Обязательные переменные
JWT_SECRET=your-super-secret-jwt-key-here-minimum-32-characters
DB_HOST=localhost
DB_PORT=5432
DB_USER=bankapp
DB_PASSWORD=bankapp123
DB_NAME=bankapp

# SMTP настройки (опционально)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=noreply@example.com
SMTP_PASSWORD=smtp_password

# PGP ключи для шифрования карт
PGP_PUBLIC_KEY="-----BEGIN PGP PUBLIC KEY BLOCK-----\n...\n-----END PGP PUBLIC KEY BLOCK-----"
PGP_PRIVATE_KEY="-----BEGIN PGP PRIVATE KEY BLOCK-----\n...\n-----END PGP PRIVATE KEY BLOCK-----"
```

5. **Скомпилируйте приложение:**
```bash
go build -o learn-bank-app ./cmd/app
```

6. **Запустите приложение:**
```bash
source .env && ./learn-bank-app
```

Приложение будет доступно на `http://localhost:8080`

## API Endpoints

### Аутентификация (Публичные endpoints)

#### Регистрация пользователя
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "SecurePass123!"
}
```

**Ответ:**
```json
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user_id": "1",
    "email": "john@example.com",
    "expires_at": "2025-01-01T12:00:00Z"
  },
  "success": true
}
```

#### Авторизация
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "SecurePass123!"
}
```

**Ответ:**
```json
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  },
  "success": true
}
```

### Управление счетами (Требуют авторизации)

*Все защищенные endpoints требуют заголовок:*
```
Authorization: Bearer YOUR_JWT_TOKEN
```

#### Создание счета
```http
POST /api/v1/accounts
Content-Type: application/json

{
  "name": "Основной счет",
  "account_type": "checking"
}
```

**Ответ:**
```json
{
  "data": {
    "id": "1",
    "user_id": "1",
    "account_number": "40817810008780678799",
    "name": "Основной счет",
    "account_type": "checking",
    "balance": 0,
    "currency": "RUB",
    "status": "active",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  },
  "success": true
}
```

#### Получение списка счетов
```http
GET /api/v1/accounts
```

**Ответ:**
```json
{
  "data": [
    {
      "id": "1",
      "user_id": "1",
      "account_number": "40817810008780678799",
      "name": "Основной счет",
      "account_type": "checking",
      "balance": 1000.50,
      "currency": "RUB",
      "status": "active",
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ],
  "success": true
}
```

#### Пополнение счета
```http
POST /api/v1/accounts/{account_id}/deposit
Content-Type: application/json

{
  "amount": 1000.50
}
```

**Ответ:**
```json
{
  "data": {
    "message": "Deposit successful"
  },
  "success": true
}
```

#### Списание со счета
```http
POST /api/v1/accounts/{account_id}/withdraw
Content-Type: application/json

{
  "amount": 500.00
}
```

**Ответ:**
```json
{
  "data": {
    "message": "Withdrawal successful"
  },
  "success": true
}
```

#### Перевод между счетами
```http
POST /api/v1/transfer
Content-Type: application/json

{
  "from_account_id": "1",
  "to_account_id": "2",
  "amount": 250.00
}
```

**Ответ:**
```json
{
  "data": {
    "message": "Transfer successful"
  },
  "success": true
}
```

### Управление картами

#### Выпуск новой карты
```http
POST /api/v1/cards
Content-Type: application/json

{
  "account_id": "1",
  "card_type": "debit",
  "pin_code": "1234",
  "daily_limit": 10000,
  "monthly_limit": 100000
}
```

**Ответ:**
```json
{
  "data": {
    "id": "1",
    "account_id": "1",
    "masked_number": "****-****-****-XXXX",
    "card_type": "LEARNBANK",
    "expiry_month": 6,
    "expiry_year": 2028,
    "status": "active",
    "daily_limit": 100000,
    "monthly_limit": 1000000,
    "daily_spent": 0,
    "monthly_spent": 0,
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  },
  "success": true
}
```

#### Получение карт по счету
```http
GET /api/v1/accounts/{account_id}/cards
```

**Ответ:**
```json
{
  "data": {
    "cards": [
      {
        "id": "1",
        "account_id": "1",
        "masked_number": "****-****-****-XXXX",
        "card_type": "LEARNBANK",
        "expiry_month": 6,
        "expiry_year": 2028,
        "status": "active",
        "daily_limit": 100000,
        "monthly_limit": 1000000,
        "daily_spent": 0,
        "monthly_spent": 0,
        "created_at": "2024-01-01T12:00:00Z",
        "updated_at": "2024-01-01T12:00:00Z"
      }
    ],
    "count": 1
  },
  "success": true
}
```

#### Оплата картой
```http
POST /api/v1/cards/{card_id}/payment
Content-Type: application/json

{
  "amount": 150.00,
  "merchant_id": "SHOP123",
  "cvv": "123"
}
```

**Ответ:**
```json
{
  "data": {
    "message": "Payment successful"
  },
  "success": true
}
```

### Кредитные операции

#### Оформление кредита
```http
POST /api/v1/credits
Content-Type: application/json

{
  "account_id": "1",
  "amount": 100000.00,
  "term_months": 12
}
```

**Ответ:**
```json
{
  "data": {
    "id": "3",
    "account_id": "1",
    "amount": 100000,
    "interest_rate": 31,
    "term_months": 12,
    "monthly_payment": 9797.97,
    "remaining_debt": 100000,
    "status": "active",
    "created_at": "2025-06-16T02:21:55.313974+05:00",
    "updated_at": "2025-06-16T02:21:55.313974+05:00"
  },
  "success": true
}
```

#### Получение графика платежей
```http
GET /api/v1/credits/{credit_id}/schedule
```

**Ответ:**
```json
{
  "data": [
    {
      "payment_number": 1,
      "payment_date": "2025-07-16",
      "total_payment": 9797.97,
      "principal_payment": 7131.30,
      "interest_payment": 2666.67,
      "remaining_debt": 92868.70
    }
    // ... остальные 11 платежей
  ],
  "success": true
}
```

### Интеграция с ЦБ РФ

#### Получение ключевой ставки ЦБ РФ
```http
GET /api/v1/cbr/rate
```

**Ответ:**
```json
{
  "data": {
    "rate": 26,
    "source": "Central Bank of Russia"
  },
  "success": true
}
```

### Аналитика

#### Месячная статистика
```http
GET /api/v1/analytics/monthly?year=2025&month=6
```

**Ответ:**
```json
{
  "data": {
    "income": 50000,
    "expenses": 35000,
    "balance": 15000,
    "month": 6,
    "year": 2025
  },
  "success": true
}
```

#### Кредитная нагрузка
```http
GET /api/v1/analytics/credit-load
```

#### Прогноз баланса
```http
POST /api/v1/analytics/balance-prediction
Content-Type: application/json

{
  "account_id": "1",
  "days": 30
}
```

## Тестирование

### Запуск тестов
```bash
# Все тесты
go test ./...

# Конкретный пакет
go test ./internal/service -v

# Покрытие
go test -cover ./...
```

### Примеры тестирования API

#### 1. Регистрация и авторизация
```bash
# Регистрация
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"SecurePass123!"}'

# Авторизация
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"SecurePass123!"}'
```

#### 2. Создание счета
```bash
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"name":"Основной счет","account_type":"checking"}'
```

#### 3. Пополнение счета
```bash
curl -X POST http://localhost:8080/api/v1/accounts/1/deposit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"amount":1000.50}'
```

#### 4. Создание кредита
```bash
curl -X POST http://localhost:8080/api/v1/credits \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"account_id":"1","amount":100000.00,"term_months":12}'
```

#### 5. Получение ключевой ставки ЦБ РФ
```bash
curl -X GET http://localhost:8080/api/v1/cbr/rate
```

### Последовательность тестирования

Рекомендуемый порядок выполнения запросов:
1. Register → Login → Create Account → Deposit → Create Card
2. Сохраните полученный JWT токен после логина для использования в последующих запросах

## Функциональные особенности

### Безопасность
- **JWT токены** с временем жизни 24 часа
- **Хеширование паролей** с использованием bcrypt
- **Шифрование данных карт** с помощью PGP
- **HMAC проверка целостности** для критичных данных
- **Проверка прав доступа** к ресурсам пользователя

### Интеграции
- **ЦБ РФ SOAP API** для получения ключевой ставки
- **SMTP** для отправки email уведомлений
- **Автоматический шедулер** для обработки просроченных платежей

### Алгоритмы
- **Алгоритм Луна** для генерации валидных номеров карт
- **Аннуитетные платежи** для расчета кредитов
- **Прогнозирование баланса** с учетом запланированных операций



## Технологии и библиотеки

- **Go 1.22+** - язык программирования
- **PostgreSQL 17** - база данных
- **jackc/pgx/v5** - драйвер PostgreSQL
- **golang-jwt/jwt/v5** - JWT токены
- **golang.org/x/crypto** - криптография
- **go-mail/mail/v2** - отправка email
- **beevik/etree** - парсинг XML
- **google/uuid** - генерация UUID

## Структура базы данных

Полная и актуальная схема базы данных описана в миграциях в директории `internal/database/migrations/`.

### Основные таблицы:

- **users** - пользователи системы
- **accounts** - банковские счета с балансами
- **cards** - виртуальные карты с шифрованием данных
- **transactions** - история всех финансовых операций
- **credits** - кредиты и займы
- **payment_schedules** - график платежей по кредитам

### Особенности схемы:

- Все таблицы имеют автоматические триггеры для обновления `updated_at`
- Данные карт шифруются с использованием PGP + HMAC
- Множественные индексы для оптимизации запросов
- Constraint'ы для валидации данных на уровне БД
- Каскадное удаление связанных записей

## Разработка

### Локальная разработка
```bash
# Установка зависимостей
go mod download

# Запуск с hot reload (с Air)
air

# Проверка линтером
golangci-lint run

# Форматирование кода
go fmt ./...
```

### Переменные окружения
Создайте `.env` файл на основе `.env.example` и заполните все необходимые значения.

## Чеклист выполненных работ

### ✅ Реализация слоя моделей
- ✅ Определение структур данных (соответствие таблицам БД)
- ✅ Сериализация/десериализация (теги JSON)
- ✅ Базовая валидация полей (email, username)
- ✅ Проверка уникальности (email, username)
- ✅ Полная валидация всех полей

### ✅ Реализация слоя репозиториев
- ✅ Инкапсуляция SQL-запросов
- ✅ Параметризованные запросы
- ✅ Простейшая обработка ошибок БД
- ✅ Управление транзакциями
- ✅ Обработка сложных ошибок БД

### ✅ Реализация слоя сервисов
- ✅ Регистрация и аутентификация
- ✅ Создание счетов, пополнение баланса
- ✅ Переводы между счетами
- ✅ Генерация карт (алгоритм Луна)
- ✅ Кредиты: расчет аннуитетных платежей
- ✅ Интеграция с SMTP (уведомления)
- ✅ Интеграция с ЦБ РФ (SOAP)
- ✅ Шедулер для списания платежей
- ✅ Логирование через slog

### ✅ Реализация слоя обработчиков
- ✅ Валидация входных данных
- ✅ Формирование HTTP-ответов (JSON)
- ✅ Вызов методов сервисов
- ✅ Реализация всех эндпоинтов из ТЗ
- ✅ Проверка прав доступа к ресурсам

### ✅ Реализация маршрутизации
- ✅ Публичные эндпоинты (/register, /login)
- ✅ Защищенные эндпоинты (/accounts, /transfer и другие)

### ✅ Реализация Middleware
- ✅ Проверка JWT-токенов
- ✅ Блокировка неавторизованных запросов
- ✅ Добавление ID пользователя в контекст

### ✅ Безопасность
- ✅ Хеширование паролей (bcrypt)
- ✅ Шифрование данных карт (PGP + HMAC)
- ✅ Хеширование CVV (bcrypt)
- ✅ Проверка прав доступа к счетам

### ✅ База данных
- ✅ Создание минимальных таблиц
