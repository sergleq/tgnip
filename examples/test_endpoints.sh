#!/bin/bash

# Тестирование эндпойнтов webhook сервера

echo "Тестирование webhook сервера..."

# Ждем немного для запуска сервера
sleep 2

# Тест health endpoint
echo "1. Тестирование health endpoint..."
curl -s http://localhost:8080/healthz | jq . 2>/dev/null || curl -s http://localhost:8080/healthz
echo -e "\n"

# Тест webhook endpoint без секретного токена
echo "2. Тестирование webhook endpoint (без токена)..."
curl -s -X POST http://localhost:8080/telegram/webhook \
  -H "Content-Type: application/json" \
  -d '{"message":{"chat":{"id":123},"text":"test message"}}'
echo -e "\n"

# Тест webhook endpoint с неправильным токеном
echo "3. Тестирование webhook endpoint (с неправильным токеном)..."
curl -s -X POST http://localhost:8080/telegram/webhook \
  -H "Content-Type: application/json" \
  -H "X-Telegram-Bot-Api-Secret-Token: wrong_token" \
  -d '{"message":{"chat":{"id":123},"text":"test message"}}'

# Тест webhook endpoint с правильным токеном (если установлен)
echo "4. Тестирование webhook endpoint (с правильным токеном)..."
SECRET_TOKEN="${WEBHOOK_SECRET_TOKEN:-test_secret}"
curl -s -X POST http://localhost:8080/telegram/webhook \
  -H "Content-Type: application/json" \
  -H "X-Telegram-Bot-Api-Secret-Token: $SECRET_TOKEN" \
  -d '{"message":{"chat":{"id":123},"text":"test message"}}'
echo -e "\n"

# Тест несуществующего endpoint
echo "5. Тестирование несуществующего endpoint..."
curl -s http://localhost:8080/nonexistent
echo -e "\n"

echo "Тестирование завершено!"
