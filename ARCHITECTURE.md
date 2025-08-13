# Архитектура TGNIP

## Компоненты

### main.go
- Точка входа приложения
- Инициализация логгера и переменных окружения
- Обработка Telegram сообщений
- Координация между компонентами

### content_extractor.go
- HTTP запросы к веб-страницам
- Извлечение контента с помощью go-readability
- Парсинг метаданных (заголовок, автор, дата)
- Обработка HTML таблиц и блоков кода
- Валидация URL

### markdown_converter.go
- Конвертация HTML в Markdown
- Создание структурированного документа
- Экранирование специальных символов
- Генерация имен файлов

### localization.go
- Определение языка пользователя
- Локализация сообщений
- Поддержка множественных языков

## Поток данных

1. Telegram API → main.go
2. main.go → content_extractor.go (валидация URL)
3. content_extractor.go → HTTP запрос → go-readability
4. content_extractor.go → markdown_converter.go
5. main.go → Telegram API (отправка файла)

## Технологии

- **Go 1.21+** - основной язык
- **go-telegram-bot-api** - Telegram Bot API
- **go-readability** - извлечение контента
- **html-to-markdown** - конвертация в Markdown
- **goquery** - парсинг HTML
- **logrus** - логирование
- **Docker** - контейнеризация
