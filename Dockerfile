# Используем официальный образ Go
FROM golang:1.21-alpine AS builder

# Устанавливаем необходимые пакеты
RUN apk add --no-cache git

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Финальный образ
FROM alpine:latest

# Устанавливаем ca-certificates для HTTPS запросов
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарный файл из builder
COPY --from=builder /app/main .

# Копируем пример файла конфигурации
COPY --from=builder /app/env.example .env.example

# Запускаем приложение
CMD ["./main"]
