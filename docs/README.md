# Документация TGNIP

Добро пожаловать в документацию проекта TGNIP - Telegram бота для преобразования веб-страниц в Markdown файлы.

## Структура документации

### 📚 Основная документация
- [README](../README.md) - Основная информация о проекте
- [CHANGELOG](CHANGELOG.md) - История изменений
- [CONTRIBUTING](CONTRIBUTING.md) - Руководство по участию в разработке
- [ARCHITECTURE](ARCHITECTURE.md) - Архитектура проекта

### 🛠️ Руководства (Guides)
- [BLOCKING_BYPASS_GUIDE](guides/BLOCKING_BYPASS_GUIDE.md) - Обход блокировок
- [RATE_LIMITING_GUIDE](guides/RATE_LIMITING_GUIDE.md) - Ограничение частоты запросов
- [WEBHOOK_SETUP](guides/WEBHOOK_SETUP.md) - Настройка Webhook
- [ETHICAL_SCRAPING](guides/ETHICAL_SCRAPING.md) - Этичный веб-скрапинг

### 📖 API документация
- [ETHICAL_SCRAPING_SUMMARY](api/ETHICAL_SCRAPING_SUMMARY.md) - Этичный скрапинг
- [BLOCKING_BYPASS_SUMMARY](api/BLOCKING_BYPASS_SUMMARY.md) - Обход блокировок
- [ENCODING_SUMMARY](api/ENCODING_SUMMARY.md) - Кодировки
- [INTEGRATION_SUMMARY](api/INTEGRATION_SUMMARY.md) - Интеграция
- [LANGUAGE_DETECTION_SUMMARY](api/LANGUAGE_DETECTION_SUMMARY.md) - Определение языка
- [RATE_LIMITING_SUMMARY](api/RATE_LIMITING_SUMMARY.md) - Ограничение частоты
- [WEBHOOK](api/WEBHOOK.md) - Webhook API

### 💡 Примеры (Examples)
- [LOCALIZATION_EXAMPLES](examples/LOCALIZATION_EXAMPLES.md) - Примеры локализации
- [README_INTEGRATION](examples/README_INTEGRATION.md) - Примеры интеграции
- [README_LANGUAGE_EXTRACTION](examples/README_LANGUAGE_EXTRACTION.md) - Извлечение языка

### 🔧 Техническая документация
- [ENCODING_FIX](ENCODING_FIX.md) - Исправления кодировок
- [LANGUAGE_DETECTION_README](LANGUAGE_DETECTION_README.md) - Определение языка
- [LANGUAGE_DETECTION_REMOVAL](LANGUAGE_DETECTION_REMOVAL.md) - Удаление определения языка
- [PARALLELISM](PARALLELISM.md) - Параллелизм
- [REFACTORING_REPORT](REFACTORING_REPORT.md) - Отчет о рефакторинге
- [VERSION](VERSION) - Версия проекта

## Быстрый старт

1. Установите зависимости: `go mod download`
2. Настройте переменные окружения (см. `env.example`)
3. Запустите бота: `go run main.go`

## Структура проекта

```
tgnip/
├── cmd/                    # Дополнительные команды
│   ├── delete_webhook/     # Удаление webhook
│   └── integration_example/ # Пример интеграции
├── docs/                   # Документация
│   ├── guides/            # Руководства
│   ├── api/               # API документация
│   └── examples/          # Примеры
├── internal/              # Внутренние пакеты
│   ├── config.go          # Конфигурация
│   ├── content_extractor.go # Извлечение контента
│   ├── ethical_scraper.go # Этичный скрапинг
│   ├── handlers.go        # Обработчики сообщений
│   ├── language_detector.go # Определение языка
│   ├── localization.go    # Локализация
│   ├── markdown_converter.go # Конвертация в Markdown
│   └── webhook.go         # Webhook сервер
├── pkg/                   # Публичные пакеты (пока пусто)
├── scripts/               # Скрипты (пока пусто)
├── examples/              # Примеры использования
├── main.go               # Главный файл
├── go.mod                # Зависимости Go
├── go.sum                # Хеши зависимостей
├── Dockerfile            # Docker образ
├── docker-compose.yml    # Docker Compose
├── Makefile              # Команды сборки
├── .env.example          # Пример переменных окружения
└── README.md             # Основная документация
```

## Поддержка

Если у вас есть вопросы или проблемы, создайте issue в репозитории проекта.
