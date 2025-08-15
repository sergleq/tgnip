#!/bin/bash

# Пример запуска бота в webhook режиме

# Установите ваш токен бота
export TELEGRAM_BOT_TOKEN="your_bot_token_here"

# Установите URL для webhook (должен быть HTTPS в продакшене)
export WEBHOOK_URL="https://your-domain.com"

# Порт для webhook сервера (по умолчанию 8080)
export WEBHOOK_PORT="8080"

# Уровень логирования
export LOG_LEVEL="info"

echo "Запуск бота в webhook режиме..."
echo "Webhook URL: $WEBHOOK_URL"
echo "Порт: $WEBHOOK_PORT"

# Запуск бота
go run main.go
