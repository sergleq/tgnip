# Telegram Bot для преобразования ссылок в Markdown

Бот для автоматического извлечения контента из веб-страниц и преобразования в markdown файлы с поддержкой локализации и безопасного webhook режима.

**[🚀 Запустить бота](https://t.me/UrlToMarkdown_bot)**

## 🚀 Возможности

### ✅ Основной функционал:
- **Извлечение контента** из веб-страниц с использованием gocolly/colly
- **Конвертация в Markdown** с сохранением форматирования
- **Поддержка кода** с определением языка программирования
- **Извлечение метаданных** (заголовок, автор, дата)
- **Локализация** на русском и английском языках
- **Fallback механизм** для надежного извлечения данных

### ✅ Производительность:
- **Rate Limiting** для Telegram API
- **Метрики производительности** в реальном времени
- **Graceful handling** перегрузки

### ✅ Безопасность:
- **Webhook-only режим** - работает только через webhook
- **Обязательный секретный токен** - защита от несанкционированных запросов
- **HTTPS поддержка** - безопасная передача данных
- **Валидация запросов** - проверка заголовков и JSON данных

### ✅ Локализация:
- **Автоматическое определение языка** пользователя
- **Локализованные markdown файлы** на языке пользователя
- **Fallback на английский** для неизвестных языков

## 📦 Установка

### Требования:
- Go 1.23+ (для поддержки gocolly/colly v2)
- Telegram Bot Token
- HTTPS домен для webhook (в продакшене)

### Сборка:
```bash
git clone https://github.com/sergleq/tgnip.git
cd tgnip
go build -o tgnip .
```

### Docker:
```bash
docker build -t tgnip .
docker run -e TELEGRAM_BOT_TOKEN=your_token -e WEBHOOK_URL=https://your-domain.com -e WEBHOOK_SECRET_TOKEN=your_secret tgnip
```

## ⚙️ Конфигурация

### Обязательные переменные окружения:
```bash
# Telegram Bot
TELEGRAM_BOT_TOKEN=your_bot_token_here

# Webhook (обязательно)
WEBHOOK_URL=https://your-domain.com
WEBHOOK_SECRET_TOKEN=your_secret_token_here

# Порт webhook сервера (по умолчанию: 8080)
WEBHOOK_PORT=8080
```

### Опциональные настройки:
```bash
# HTTP клиент
HTTP_TIMEOUT=30                   # Таймаут HTTP запросов в секундах
MAX_RETRIES=3                     # Максимальное количество повторов

# Логирование
LOG_LEVEL=info                    # debug, info, warn, error

# SSL сертификаты (для HTTPS)
SSL_CERT_FILE=/path/to/cert.pem
SSL_KEY_FILE=/path/to/key.pem
```

## 🚀 Запуск

### Локальная разработка:
```bash
# Установите переменные окружения
export TELEGRAM_BOT_TOKEN="your_bot_token"
export WEBHOOK_URL="https://your-domain.com"
export WEBHOOK_SECRET_TOKEN="your_secret_token"

# Запустите бота
./tgnip
```

### Разработка с ngrok:
```bash
# 1. Запустите ngrok
ngrok http 8080

# 2. Установите переменные окружения
export TELEGRAM_BOT_TOKEN="your_bot_token"
export WEBHOOK_URL="https://your-ngrok-url.ngrok.io"
export WEBHOOK_SECRET_TOKEN="your_secret_token"

# 3. Запустите бота
./tgnip

# 4. Установите webhook вручную
curl -X POST https://api.telegram.org/bot<TOKEN>/setWebhook \
  -H 'Content-Type: application/json' \
  -d '{"url": "https://your-ngrok-url.ngrok.io/telegram/webhook", "secret_token": "your_secret_token"}'
```

## 🚀 Деплой на Fly.io

### Быстрый деплой:
```bash
# 1. Настройте секреты
./scripts/setup_fly_secrets.sh

# 2. Запустите деплой
./scripts/deploy.sh

# 3. Установите webhook
./scripts/setup_webhook.sh
```

### Ручной деплой:
```bash
# Установите секреты
flyctl secrets set TELEGRAM_BOT_TOKEN="your_bot_token"
flyctl secrets set WEBHOOK_URL="https://tgnip.fly.dev"
flyctl secrets set WEBHOOK_SECRET_TOKEN="your_secret_token"

# Деплой
flyctl deploy

# Установите webhook
./scripts/setup_webhook.sh
```

Подробная документация: [README_DEPLOY.md](README_DEPLOY.md)

## 📊 Мониторинг

### Логи:
```
INFO[2025-08-15T09:12:04Z] Переменные окружения загружены из .env файла
INFO[2025-08-15T09:12:04Z] Бот UrlToMarkdown_bot запущен
INFO[2025-08-15T09:12:04Z] Запуск webhook сервера
INFO[2025-08-15T09:12:04Z] Используем WEBHOOK_SECRET_TOKEN из переменной окружения
INFO[2025-08-15T09:12:04Z] HTTP сервер запущен на 0.0.0.0:8080
```

### Health check:
```bash
curl https://your-domain.com/healthz
```

### Проверка webhook:
```bash
curl https://api.telegram.org/bot<TOKEN>/getWebhookInfo
```

## 🔒 Безопасность

### Webhook защита:
- **Обязательный секретный токен** - бот не запустится без `WEBHOOK_SECRET_TOKEN`
- **Проверка заголовков** - валидация `X-Telegram-Bot-Api-Secret-Token`
- **Валидация JSON** - проверка структуры входящих данных
- **HTTPS только** - поддержка SSL/TLS в продакшене

### Идемпотентная установка webhook:
- Webhook устанавливается только вручную
- Всегда с одним и тем же секретным токеном
- Не перезаписывается при перезапуске

### Переменные окружения:
- Приоритет у переменных окружения над .env файлом
- Секреты хранятся в fly.io secrets
- Логирование источника переменных

## 📁 Структура проекта

```
tgnip/
├── internal/              # Внутренние пакеты
│   ├── config.go          # Конфигурация
│   ├── content_extractor.go # Извлечение контента
│   ├── ethical_scraper.go # Этичный скрапинг
│   ├── handlers.go        # Обработчики сообщений
│   ├── language_detector.go # Определение языка
│   ├── localization.go    # Локализация
│   ├── markdown_converter.go # Конвертация в Markdown
│   └── webhook.go         # Webhook сервер
├── scripts/               # Скрипты автоматизации
│   ├── setup_fly_secrets.sh # Настройка секретов fly.io
│   ├── deploy.sh          # Деплой на fly.io
│   └── setup_webhook.sh   # Установка webhook
├── examples/              # Примеры использования
│   ├── webhook_example.sh # Пример запуска webhook
│   └── test_endpoints.sh  # Тестирование эндпойнтов
├── main.go               # Главный файл
├── go.mod                # Зависимости Go
├── Dockerfile            # Docker образ
├── docker-compose.yml    # Docker Compose
├── fly.toml             # Конфигурация Fly.io
├── Makefile              # Команды сборки
└── README.md             # Основная документация
```

## 🔧 Использование

### Команды бота:
- `/start` - приветственное сообщение

### Обработка ссылок:
1. Отправьте ссылку на веб-страницу боту
2. Бот извлечет контент и создаст markdown файл
3. Файл будет отправлен с локализованными заголовками

### Примеры поддерживаемых сайтов:
- Статьи и новости
- Блог-посты
- Документация
- Технические статьи

## 🧪 Тестирование

### Запуск тестов:
```bash
# Все тесты
go test -v

# Тестирование эндпойнтов
./examples/test_endpoints.sh
```

### Тестирование webhook:
```bash
# Тест без токена (должен вернуть 401)
curl -X POST http://localhost:8080/telegram/webhook \
  -H "Content-Type: application/json" \
  -d '{"message":{"chat":{"id":123},"text":"test"}}'

# Тест с правильным токеном (должен вернуть 200 OK)
curl -X POST http://localhost:8080/telegram/webhook \
  -H "Content-Type: application/json" \
  -H "X-Telegram-Bot-Api-Secret-Token: your_secret_token" \
  -d '{"message":{"chat":{"id":123},"text":"test"}}'
```

## 🛠️ Разработка

### Структура проекта:
```
tgnip/
├── main.go                 # Основная логика и webhook сервер
├── internal/webhook.go     # Webhook сервер и безопасность
├── internal/config.go      # Конфигурация
├── internal/content_extractor.go # Извлечение контента
├── internal/markdown_converter.go # Конвертация в markdown
├── internal/localization.go # Локализация
├── scripts/               # Скрипты автоматизации
├── examples/              # Примеры и тесты
└── *.md                   # Документация
```

### Добавление новых языков:
1. Добавьте новый язык в `internal/localization.go`
2. Переведите все поля локализации
3. Добавьте тесты
4. Обновите документацию

### Безопасность при разработке:
- Всегда используйте секретный токен
- Не коммитьте секреты в репозиторий
- Используйте переменные окружения для конфиденциальных данных
- Тестируйте webhook с правильными заголовками

## 📄 Лицензия

MIT License

## �� Вклад в проект

1. Fork репозитория
2. Создайте feature branch
3. Внесите изменения
4. Добавьте тесты
5. Создайте Pull Request

## 📞 Поддержка

- Создайте Issue для багов
- Обсудите новые возможности в Discussions
- Присоединяйтесь к разработке 
