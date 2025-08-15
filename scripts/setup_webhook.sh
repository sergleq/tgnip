#!/bin/bash

# Скрипт для установки webhook с секретным токеном
# Использование: ./scripts/setup_webhook.sh

set -e

echo "🔧 Установка webhook с секретным токеном..."

# Проверяем, что flyctl установлен
if ! command -v flyctl &> /dev/null; then
    echo "❌ Ошибка: flyctl не установлен"
    exit 1
fi

# Получаем секреты из fly.io
echo "📥 Получение секретов из fly.io..."

TELEGRAM_BOT_TOKEN=$(flyctl secrets list | grep TELEGRAM_BOT_TOKEN | awk '{print $2}' || echo "")
WEBHOOK_URL=$(flyctl secrets list | grep WEBHOOK_URL | awk '{print $2}' || echo "")
WEBHOOK_SECRET_TOKEN=$(flyctl secrets list | grep WEBHOOK_SECRET_TOKEN | awk '{print $2}' || echo "")

if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
    echo "❌ Ошибка: TELEGRAM_BOT_TOKEN не найден в секретах"
    echo "Установите секреты: ./scripts/setup_fly_secrets.sh"
    exit 1
fi

if [ -z "$WEBHOOK_URL" ]; then
    echo "❌ Ошибка: WEBHOOK_URL не найден в секретах"
    echo "Установите секреты: ./scripts/setup_fly_secrets.sh"
    exit 1
fi

if [ -z "$WEBHOOK_SECRET_TOKEN" ]; then
    echo "❌ Ошибка: WEBHOOK_SECRET_TOKEN не найден в секретах"
    echo "Установите секреты: ./scripts/setup_fly_secrets.sh"
    exit 1
fi

echo "✅ Секреты получены успешно"

# Формируем URL для webhook
WEBHOOK_FULL_URL="${WEBHOOK_URL}/telegram/webhook"

echo "🔗 Устанавливаем webhook: $WEBHOOK_FULL_URL"
echo "🔐 Секретный токен: ${WEBHOOK_SECRET_TOKEN:0:8}..."

# Устанавливаем webhook через Telegram API
RESPONSE=$(curl -s -X POST "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/setWebhook" \
  -H "Content-Type: application/json" \
  -d "{
    \"url\": \"${WEBHOOK_FULL_URL}\",
    \"secret_token\": \"${WEBHOOK_SECRET_TOKEN}\"
  }")

echo "📡 Ответ от Telegram API:"
echo "$RESPONSE" | jq . 2>/dev/null || echo "$RESPONSE"

# Проверяем успешность установки
if echo "$RESPONSE" | grep -q '"ok":true'; then
    echo "✅ Webhook установлен успешно!"
    
    # Проверяем информацию о webhook
    echo "📋 Информация о webhook:"
    curl -s "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/getWebhookInfo" | jq .
else
    echo "❌ Ошибка установки webhook"
    exit 1
fi
