# Telegram Bot для преобразования ссылок в Markdown

Бот для автоматического извлечения контента из веб-страниц и преобразования в markdown файлы с поддержкой параллелизма, локализации и webhook.

## 🚀 Возможности

### ✅ Основной функционал:
- **Извлечение контента** из веб-страниц с использованием gocolly/colly
- **Конвертация в Markdown** с сохранением форматирования
- **Поддержка кода** с определением языка программирования
- **Извлечение метаданных** (заголовок, автор, дата)
- **Локализация** на русском и английском языках
- **Fallback механизм** для надежного извлечения данных

### ✅ Производительность:
- **Worker Pool** для параллельной обработки запросов
- **Rate Limiting** для Telegram API
- **Метрики производительности** в реальном времени
- **Graceful handling** перегрузки

### ✅ Режим работы:
- **Webhook** - для продакшена с высокой нагрузкой

### ✅ Локализация:
- **Автоматическое определение языка** пользователя
- **Локализованные markdown файлы** на языке пользователя
- **Fallback на английский** для неизвестных языков

## 📦 Установка

### Требования:
- Go 1.23+ (для поддержки gocolly/colly v2)
- Telegram Bot Token

### Сборка:
```bash
git clone https://github.com/sergleq/tgnip.git
cd tgnip
go build -o tgnip .
```

### Docker:
```bash
docker build -t tgnip .
docker run -e TELEGRAM_BOT_TOKEN=your_token tgnip
```

## ⚙️ Конфигурация

### Базовые настройки:
```bash
# Обязательные
TELEGRAM_BOT_TOKEN=your_bot_token_here

# HTTP клиент
HTTP_TIMEOUT=30                   # Таймаут HTTP запросов в секундах
MAX_RETRIES=3                     # Максимальное количество повторов

# Логирование
LOG_LEVEL=info                    # debug, info, warn, error
```

### Webhook настройки (обязательно):
```bash
# Webhook режим
WEBHOOK_URL=https://your-domain.com  # URL для webhook (обязательно)
WEBHOOK_PORT=8080                     # Порт для webhook сервера
SSL_CERT_FILE=/path/to/cert.pem       # SSL сертификат
SSL_KEY_FILE=/path/to/key.pem         # SSL ключ
```

## 🚀 Запуск

### Webhook режим (обязательно):
```bash
WEBHOOK_URL=https://your-domain.com ./tgnip
```

### Разработка с ngrok:
```bash
# 1. Запустите ngrok
ngrok http 8080

# 2. Запустите бота с полученным URL
WEBHOOK_URL=https://your-ngrok-url.ngrok.io ./tgnip
```

## 🚀 Деплой на Fly.io

### Быстрый деплой:
```bash
# 1. Настройте секреты
./scripts/setup_fly_secrets.sh

# 2. Запустите деплой
./scripts/deploy.sh
```

Подробная документация: [README_DEPLOY.md](README_DEPLOY.md)

## 📊 Мониторинг

### Логи:
```
INFO[2025-08-13T15:21:00Z] Бот UrlToMarkdown_bot запущен
INFO[2025-08-13T15:21:00Z] Запуск в polling режиме
INFO[2025-08-13T15:21:05Z] Обрабатывается сообщение от пользователя 123456
```

### Health check (webhook режим):
```bash
curl https://your-domain.com/health
```

## 📁 Структура проекта

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
├── pkg/                   # Публичные пакеты
├── scripts/               # Скрипты
├── examples/              # Примеры использования
├── main.go               # Главный файл
├── go.mod                # Зависимости Go
├── Dockerfile            # Docker образ
├── docker-compose.yml    # Docker Compose
├── Makefile              # Команды сборки
└── README.md             # Основная документация
```

## 📚 Документация

Подробная документация доступна в папке [docs/](docs/README.md):

- [Руководства](docs/guides/) - Пошаговые инструкции
- [API документация](docs/api/) - Техническая документация
- [Примеры](docs/examples/) - Примеры использования
- [Интеграция Colly](docs/COLLY_INTEGRATION.md) - Документация по gocolly/colly

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
```

## 📚 Документация

- [PARALLELISM.md](PARALLELISM.md) - Документация по параллелизму
- [WEBHOOK.md](WEBHOOK.md) - Документация по webhook
- [LOCALIZATION_EXAMPLES.md](LOCALIZATION_EXAMPLES.md) - Примеры локализации
- [REFACTORING_REPORT.md](REFACTORING_REPORT.md) - Отчет о рефакторинге
- [COLLY_INTEGRATION.md](docs/COLLY_INTEGRATION.md) - Интеграция gocolly/colly

## 🔒 Безопасность

### Webhook:
- Автоматическая генерация секретного токена
- Проверка заголовков запросов
- Валидация JSON данных
- SSL/TLS поддержка

### Обработка ошибок:
- Graceful handling перегрузки
- Fallback механизмы
- Детальное логирование

## 🛠️ Разработка

### Структура проекта:
```
tgnip/
├── main.go                 # Основная логика и worker pool
├── webhook.go              # Webhook сервер
├── config.go               # Конфигурация
├── content_extractor.go    # Извлечение контента
├── markdown_converter.go   # Конвертация в markdown
├── localization.go         # Локализация
├── *_test.go              # Тесты
└── *.md                   # Документация
```

### Добавление новых языков:
1. Добавьте новый язык в `localization.go`
2. Переведите все поля локализации
3. Добавьте тесты
4. Обновите документацию

## 📄 Лицензия

MIT License

## 🤝 Вклад в проект

1. Fork репозитория
2. Создайте feature branch
3. Внесите изменения
4. Добавьте тесты
5. Создайте Pull Request

## 📞 Поддержка

- Создайте Issue для багов
- Обсудите новые возможности в Discussions
- Присоединяйтесь к разработке 