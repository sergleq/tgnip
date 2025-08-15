# Настройка Webhook Сервера

## Описание

Бот работает только в webhook режиме с HTTP-сервером, который слушает на `0.0.0.0:8080` и обрабатывает запросы на эндпойнте `/telegram/webhook`.

## Конфигурация

### Переменные окружения

- `TELEGRAM_BOT_TOKEN` - токен вашего Telegram бота (обязательно)
- `WEBHOOK_URL` - URL для webhook (обязательно)
- `WEBHOOK_PORT` - порт для webhook сервера (по умолчанию: 8080)
- `WEBHOOK_SECRET_TOKEN` - секретный токен для проверки webhook запросов (опционально)
- `SSL_CERT_FILE` - путь к SSL сертификату (для HTTPS)
- `SSL_KEY_FILE` - путь к SSL ключу (для HTTPS)

### Режим работы

Бот работает только в webhook режиме. Переменная `WEBHOOK_URL` обязательна для запуска.

## Запуск

### Локальная разработка

1. Установите переменные окружения:
```bash
export TELEGRAM_BOT_TOKEN="your_bot_token"
export WEBHOOK_URL="https://your-domain.com"
export WEBHOOK_PORT="8080"
export WEBHOOK_SECRET_TOKEN="your_secret_token"
```

2. Запустите бота:
```bash
go run main.go
```

### С Docker

1. Создайте файл `.env`:
```bash
TELEGRAM_BOT_TOKEN=your_bot_token
WEBHOOK_URL=https://your-domain.com
WEBHOOK_PORT=8080
```

2. Запустите с Docker Compose:
```bash
docker-compose up -d
```

### Для разработки с ngrok

1. Установите ngrok: https://ngrok.com/

2. Запустите туннель:
```bash
ngrok http 8080
```

3. Скопируйте HTTPS URL из ngrok и установите переменную:
```bash
export WEBHOOK_URL="https://your-ngrok-url.ngrok.io"
```

4. Запустите бота:
```bash
go run main.go
```

## Эндпойнты

### `/telegram/webhook`
- **Метод**: POST
- **Content-Type**: application/json
- **Описание**: Обрабатывает входящие webhook запросы от Telegram

### `/healthz`
- **Метод**: GET
- **Описание**: Health check эндпойнт
- **Ответ**: JSON с статусом и URL webhook

## Примеры

### Тестирование health check
```bash
curl http://localhost:8080/healthz
```

### Тестирование webhook (локально)
```bash
curl -X POST http://localhost:8080/telegram/webhook \
  -H "Content-Type: application/json" \
  -d '{"message":{"chat":{"id":123},"text":"test"}}'
```

## Безопасность

- Сервер генерирует секретный токен для проверки webhook запросов
- Рекомендуется использовать HTTPS в продакшене
- Для разработки можно использовать HTTP с ngrok

## Логирование

Бот логирует все входящие webhook запросы и ошибки. Уровень логирования можно настроить через переменную `LOG_LEVEL`:
- `debug` - подробное логирование
- `info` - стандартное логирование (по умолчанию)
- `warn` - только предупреждения и ошибки
- `error` - только ошибки
