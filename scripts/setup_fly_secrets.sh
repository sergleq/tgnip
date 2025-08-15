#!/bin/bash

# Скрипт для настройки секретов fly.io
# Использование: ./scripts/setup_fly_secrets.sh

echo "Настройка секретов для fly.io..."

# Проверяем, что flyctl установлен
if ! command -v flyctl &> /dev/null; then
    echo "Ошибка: flyctl не установлен. Установите flyctl: https://fly.io/docs/hands-on/install-flyctl/"
    exit 1
fi

# Проверяем, что мы в правильной директории
if [ ! -f "fly.toml" ]; then
    echo "Ошибка: fly.toml не найден. Запустите скрипт из корневой директории проекта."
    exit 1
fi

# Запрашиваем токен бота
echo -n "Введите TELEGRAM_BOT_TOKEN: "
read -s TELEGRAM_BOT_TOKEN
echo

if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
    echo "Ошибка: TELEGRAM_BOT_TOKEN не может быть пустым"
    exit 1
fi

	# Запрашиваем URL для webhook
	echo -n "Введите WEBHOOK_URL (например, https://tgnip.fly.dev): "
	read WEBHOOK_URL

	if [ -z "$WEBHOOK_URL" ]; then
		echo "Ошибка: WEBHOOK_URL не может быть пустым"
		exit 1
	fi

	# Запрашиваем секретный токен для webhook
	echo -n "Введите WEBHOOK_SECRET_TOKEN (или нажмите Enter для автогенерации): "
	read WEBHOOK_SECRET_TOKEN

	if [ -z "$WEBHOOK_SECRET_TOKEN" ]; then
		WEBHOOK_SECRET_TOKEN=$(openssl rand -hex 32)
		echo "Сгенерирован секретный токен: $WEBHOOK_SECRET_TOKEN"
	fi

	# Устанавливаем секреты
	echo "Установка секретов..."

	flyctl secrets set TELEGRAM_BOT_TOKEN="$TELEGRAM_BOT_TOKEN"
	flyctl secrets set WEBHOOK_URL="$WEBHOOK_URL"
	flyctl secrets set WEBHOOK_SECRET_TOKEN="$WEBHOOK_SECRET_TOKEN"

# Опциональные секреты
echo -n "Установить дополнительные секреты? (y/n): "
read -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -n "Введите LOG_LEVEL (по умолчанию: info): "
    read LOG_LEVEL
    if [ -n "$LOG_LEVEL" ]; then
        flyctl secrets set LOG_LEVEL="$LOG_LEVEL"
    fi

    echo -n "Введите HTTP_TIMEOUT (по умолчанию: 30): "
    read HTTP_TIMEOUT
    if [ -n "$HTTP_TIMEOUT" ]; then
        flyctl secrets set HTTP_TIMEOUT="$HTTP_TIMEOUT"
    fi

    echo -n "Введите MAX_RETRIES (по умолчанию: 3): "
    read MAX_RETRIES
    if [ -n "$MAX_RETRIES" ]; then
        flyctl secrets set MAX_RETRIES="$MAX_RETRIES"
    fi
fi

echo "Секреты установлены успешно!"
echo "Для просмотра секретов используйте: flyctl secrets list"
