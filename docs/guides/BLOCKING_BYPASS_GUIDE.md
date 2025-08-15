# Руководство по обходу блокировок

Этот документ описывает новые возможности бота для обхода блокировок на веб-сайтах.

## Возможности

### 1. Ротация User-Agent
Бот автоматически использует случайные User-Agent строки популярных браузеров для имитации реального пользователя.

**Доступные User-Agent по умолчанию:**
- Chrome (Windows, Mac, Linux)
- Firefox (Windows, Mac)
- Safari (Mac)
- Edge (Windows)

### 2. Поддержка прокси-серверов
Бот может использовать прокси-серверы для обхода IP-блокировок.

**Поддерживаемые типы прокси:**
- HTTP прокси
- HTTPS прокси
- SOCKS5 прокси

### 3. Механизм повторных попыток
При ошибках подключения бот автоматически повторяет запросы с настраиваемой задержкой.

### 4. Задержки между запросами
Настраиваемые задержки между запросами для имитации человеческого поведения.

### 5. Расширенные HTTP заголовки
Бот отправляет заголовки, имитирующие реальный браузер.

## Конфигурация

### Переменные окружения

Добавьте следующие переменные в ваш `.env` файл:

```bash
# HTTP Client Configuration
HTTP_TIMEOUT=30s               # Таймаут HTTP запросов
HTTP_RETRY_COUNT=3             # Количество повторных попыток
HTTP_RETRY_DELAY=2s            # Задержка между повторными попытками
HTTP_REQUEST_DELAY=1s          # Задержка между запросами
HTTP_PROXY_URL=                # URL прокси-сервера (опционально)
HTTP_USER_AGENTS=              # Пользовательские User-Agent (опционально)
```

### Примеры конфигурации

#### Базовая конфигурация
```bash
HTTP_TIMEOUT=30s
HTTP_RETRY_COUNT=3
HTTP_RETRY_DELAY=2s
HTTP_REQUEST_DELAY=1s
```

#### С прокси-сервером
```bash
HTTP_TIMEOUT=30s
HTTP_RETRY_COUNT=3
HTTP_RETRY_DELAY=2s
HTTP_REQUEST_DELAY=1s
HTTP_PROXY_URL=http://proxy.example.com:8080
```

#### С пользовательскими User-Agent
```bash
HTTP_TIMEOUT=30s
HTTP_RETRY_COUNT=3
HTTP_RETRY_DELAY=2s
HTTP_REQUEST_DELAY=1s
HTTP_USER_AGENTS="Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36,Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
```

#### Агрессивная конфигурация для сложных сайтов
```bash
HTTP_TIMEOUT=60s
HTTP_RETRY_COUNT=5
HTTP_RETRY_DELAY=5s
HTTP_REQUEST_DELAY=3s
HTTP_PROXY_URL=socks5://proxy.example.com:1080
```

## Рекомендации

### Для разных типов сайтов

#### Обычные сайты
```bash
HTTP_TIMEOUT=30s
HTTP_RETRY_COUNT=3
HTTP_RETRY_DELAY=2s
HTTP_REQUEST_DELAY=1s
```

#### Сайты с базовой защитой
```bash
HTTP_TIMEOUT=45s
HTTP_RETRY_COUNT=4
HTTP_RETRY_DELAY=3s
HTTP_REQUEST_DELAY=2s
```

#### Сайты с продвинутой защитой
```bash
HTTP_TIMEOUT=60s
HTTP_RETRY_COUNT=5
HTTP_RETRY_DELAY=5s
HTTP_REQUEST_DELAY=3s
HTTP_PROXY_URL=http://your-proxy:8080
```

### Выбор прокси-сервера

1. **HTTP прокси** - для базовой защиты
2. **HTTPS прокси** - для более надежной защиты
3. **SOCKS5 прокси** - для максимальной анонимности

### Мониторинг и отладка

Включите debug логирование для отслеживания работы:
```bash
LOG_LEVEL=debug
```

## Безопасность

### Рекомендации по безопасности

1. **Используйте надежные прокси-серверы** - избегайте бесплатных прокси
2. **Соблюдайте robots.txt** - уважайте правила сайтов
3. **Не превышайте разумные лимиты** - не перегружайте серверы
4. **Мониторьте логи** - следите за ошибками и блокировками

### Ограничения

- Бот не обходит JavaScript-защиту
- Бот не обходит капчу
- Бот не обходит биометрическую аутентификацию

## Устранение неполадок

### Частые проблемы

#### Ошибка 403 Forbidden
- Увеличьте `HTTP_REQUEST_DELAY`
- Добавьте прокси-сервер
- Проверьте User-Agent

#### Ошибка 429 Too Many Requests
- Увеличьте `HTTP_REQUEST_DELAY`
- Уменьшите `HTTP_RETRY_COUNT`
- Добавьте прокси-сервер

#### Таймауты
- Увеличьте `HTTP_TIMEOUT`
- Проверьте стабильность интернет-соединения
- Попробуйте другой прокси-сервер

### Логирование

Для диагностики проблем включите debug логирование:
```bash
LOG_LEVEL=debug
```

Это покажет:
- Используемые User-Agent
- Время выполнения запросов
- Статусы ответов
- Информацию о повторных попытках

## Примеры использования

### Docker Compose с прокси

```yaml
version: '3.8'
services:
  tgnip:
    build: .
    environment:
      - TELEGRAM_BOT_TOKEN=your_token
      - HTTP_TIMEOUT=30s
      - HTTP_RETRY_COUNT=3
      - HTTP_RETRY_DELAY=2s
      - HTTP_REQUEST_DELAY=1s
      - HTTP_PROXY_URL=http://proxy:8080
      - LOG_LEVEL=info
    depends_on:
      - proxy

  proxy:
    image: nginx:alpine
    ports:
      - "8080:8080"
    volumes:
      - ./proxy.conf:/etc/nginx/nginx.conf
```

### Kubernetes с ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tgnip-config
data:
  HTTP_TIMEOUT: "30s"
  HTTP_RETRY_COUNT: "3"
  HTTP_RETRY_DELAY: "2s"
  HTTP_REQUEST_DELAY: "1s"
  HTTP_PROXY_URL: "http://proxy-service:8080"
  LOG_LEVEL: "info"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tgnip
spec:
  replicas: 1
  selector:
    matchLabels:
      app: tgnip
  template:
    metadata:
      labels:
        app: tgnip
    spec:
      containers:
      - name: tgnip
        image: tgnip:latest
        envFrom:
        - configMapRef:
            name: tgnip-config
        env:
        - name: TELEGRAM_BOT_TOKEN
          valueFrom:
            secretKeyRef:
              name: tgnip-secrets
              key: telegram-bot-token
```
