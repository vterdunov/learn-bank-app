# Bank App

REST API для банковского сервиса на Go с использованием чистой архитектуры.

## Быстрый запуск

1. Скопируйте файл с переменными окружения:
```bash
cp .env.example .env
```

2. Отредактируйте `.env` файл и установите `JWT_SECRET`:
```bash
# Обязательно установите JWT_SECRET
JWT_SECRET=your-super-secret-jwt-key-here
```

3. Скомпилируйте приложение:
```bash
go build -o learn-bank-app cmd/app/main.go
```

4. Запустите приложение:
```bash
source .env && ./learn-bank-app
```

Приложение будет доступно на `http://localhost:8080`
