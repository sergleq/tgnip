#!/bin/bash

# Скрипт для деплоя на fly.io
# Использование: ./scripts/deploy.sh

set -e

echo "🚀 Деплой на fly.io..."

# Проверяем, что flyctl установлен
if ! command -v flyctl &> /dev/null; then
    echo "❌ Ошибка: flyctl не установлен. Установите flyctl: https://fly.io/docs/hands-on/install-flyctl/"
    exit 1
fi

# Проверяем, что мы в правильной директории
if [ ! -f "fly.toml" ]; then
    echo "❌ Ошибка: fly.toml не найден. Запустите скрипт из корневой директории проекта."
    exit 1
fi

# Проверяем, что приложение существует
if ! flyctl apps list | grep -q "tgnip"; then
    echo "❌ Ошибка: Приложение 'tgnip' не найдено в fly.io"
    echo "Создайте приложение: flyctl apps create tgnip"
    exit 1
fi

# Проверяем, что секреты установлены
echo "🔍 Проверка секретов..."
if ! flyctl secrets list | grep -q "TELEGRAM_BOT_TOKEN"; then
    echo "⚠️  Предупреждение: TELEGRAM_BOT_TOKEN не установлен"
    echo "Установите секреты: ./scripts/setup_fly_secrets.sh"
fi

	if ! flyctl secrets list | grep -q "WEBHOOK_URL"; then
		echo "⚠️  Предупреждение: WEBHOOK_URL не установлен"
		echo "Установите секреты: ./scripts/setup_fly_secrets.sh"
	fi

	if ! flyctl secrets list | grep -q "WEBHOOK_SECRET_TOKEN"; then
		echo "⚠️  Предупреждение: WEBHOOK_SECRET_TOKEN не установлен"
		echo "Установите секреты: ./scripts/setup_fly_secrets.sh"
	fi

# Собираем проект
echo "🔨 Сборка проекта..."
go build -o tgnip .

# Деплой
echo "📦 Деплой на fly.io..."
flyctl deploy

# Проверяем статус
echo "🔍 Проверка статуса..."
sleep 5
flyctl status

	# Проверяем health check
	echo "🏥 Проверка health check..."
	sleep 10
	if curl -f https://tgnip.fly.dev/healthz > /dev/null 2>&1; then
		echo "✅ Health check прошел успешно!"
	else
		echo "❌ Health check не прошел"
		echo "Проверьте логи: flyctl logs"
	fi

	echo "🎉 Деплой завершен!"
	echo "URL: https://tgnip.fly.dev"
	echo "Логи: flyctl logs"
	echo "Статус: flyctl status"
	echo ""
	echo "⚠️  ВАЖНО: Webhook не установлен автоматически!"
	echo "Установите webhook вручную: ./scripts/setup_webhook.sh"
