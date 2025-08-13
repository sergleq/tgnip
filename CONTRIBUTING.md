# Руководство по разработке

## Требования

- Go 1.21+
- Docker (опционально)
- golangci-lint

## Настройка

```bash
git clone <repo-url>
cd tgnip
go mod download
cp env.example .env
# Отредактируйте .env
```

## Команды

```bash
# Запуск
go run .

# Тесты
go test ./...

# Линтер
golangci-lint run

# Форматирование
go fmt ./...

# Сборка
go build .
```

## Стандарты кода

- Используйте `go fmt` для форматирования
- Добавляйте комментарии к экспортируемым функциям
- Всегда проверяйте ошибки
- Используйте table-driven tests
- Логируйте ошибки с помощью logrus

## Pull Request

1. Создайте ветку: `git checkout -b feature/name`
2. Реализуйте функциональность
3. Добавьте тесты
4. Запустите: `go test && golangci-lint run`
5. Создайте PR с описанием изменений
