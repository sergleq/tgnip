.PHONY: help build run test clean docker-build docker-run

BINARY_NAME=tgnip
DOCKER_IMAGE=tgnip-bot

help: ## Показать справку
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Собрать приложение
	go build -ldflags="-s -w" -o $(BINARY_NAME) .

run: ## Запустить приложение
	go run .

test: ## Запустить тесты
	go test -v ./...

clean: ## Очистить артефакты
	rm -f $(BINARY_NAME)
	go clean

docker-build: ## Собрать Docker образ
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Запустить Docker контейнер
	docker run -d --name $(DOCKER_IMAGE) --env-file .env $(DOCKER_IMAGE)

docker-stop: ## Остановить Docker контейнер
	docker stop $(DOCKER_IMAGE) || true
	docker rm $(DOCKER_IMAGE) || true

lint: ## Запустить линтер
	golangci-lint run

format: ## Форматировать код
	go fmt ./...

setup: ## Настройка проекта
	@if [ ! -f .env ]; then \
		cp env.example .env; \
		echo "Создан файл .env. Отредактируйте его и добавьте токен бота!"; \
	else \
		echo "Файл .env уже существует"; \
	fi
	go mod download
