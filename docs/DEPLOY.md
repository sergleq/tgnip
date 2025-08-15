# Деплой на Fly.io

## Предварительные требования

1. **Установите flyctl**:
   ```bash
   # macOS
   brew install flyctl
   
   # Linux
   curl -L https://fly.io/install.sh | sh
   
   # Windows
   # Скачайте с https://fly.io/docs/hands-on/install-flyctl/
   ```

2. **Авторизуйтесь в fly.io**:
   ```bash
   flyctl auth login
   ```

3. **Создайте приложение** (если еще не создано):
   ```bash
   flyctl apps create tgnip
   ```

## Быстрый деплой

### 1. Настройка секретов

Запустите интерактивный скрипт для настройки секретов:

```bash
./scripts/setup_fly_secrets.sh
```

Или установите секреты вручную:

```bash
flyctl secrets set TELEGRAM_BOT_TOKEN="your_bot_token"
flyctl secrets set WEBHOOK_URL="https://tgnip.fly.dev"
```

### 2. Деплой

Запустите скрипт деплоя:

```bash
./scripts/deploy.sh
```

Или выполните деплой вручную:

```bash
flyctl deploy
```

### 3. Установка webhook

После деплоя установите webhook вручную:

```bash
./scripts/setup_webhook.sh
```

Или установите webhook вручную через curl:

```bash
curl -X POST https://api.telegram.org/bot<TOKEN>/setWebhook \
  -H 'Content-Type: application/json' \
  -d '{"url": "https://tgnip.fly.dev/telegram/webhook", "secret_token": "<SECRET_TOKEN>"}'
```

## Ручная настройка

### Переменные окружения

Обязательные секреты:
- `TELEGRAM_BOT_TOKEN` - токен вашего Telegram бота
- `WEBHOOK_URL` - URL для webhook (например, https://tgnip.fly.dev)
- `WEBHOOK_SECRET_TOKEN` - секретный токен для проверки webhook запросов

Опциональные секреты:
- `LOG_LEVEL` - уровень логирования (debug, info, warn, error)
- `HTTP_TIMEOUT` - таймаут HTTP запросов в секундах
- `MAX_RETRIES` - максимальное количество повторных попыток

### Установка секретов

```bash
# Обязательные
flyctl secrets set TELEGRAM_BOT_TOKEN="your_bot_token"
flyctl secrets set WEBHOOK_URL="https://tgnip.fly.dev"
flyctl secrets set WEBHOOK_SECRET_TOKEN="your_secret_token"

# Опциональные
flyctl secrets set LOG_LEVEL="info"
flyctl secrets set HTTP_TIMEOUT="30"
flyctl secrets set MAX_RETRIES="3"
```

### Просмотр секретов

```bash
flyctl secrets list
```

## Управление приложением

### Статус

```bash
flyctl status
```

### Логи

```bash
# Все логи
flyctl logs

# Логи в реальном времени
flyctl logs --follow

# Логи с определенного времени
flyctl logs --since=1h
```

### Масштабирование

```bash
# Увеличить количество машин
flyctl scale count 2

# Изменить размер машины
flyctl scale vm shared-cpu-2x
```

### Перезапуск

```bash
flyctl restart
```

### Удаление

```bash
flyctl destroy
```

## Мониторинг

### Health Check

Приложение автоматически проверяется по эндпойнту `/healthz`:

```bash
curl https://tgnip.fly.dev/healthz
```

### Метрики

```bash
# Просмотр метрик
flyctl dashboard

# Мониторинг в реальном времени
flyctl dashboard --open
```

## Troubleshooting

### Проблемы с деплоем

1. **Проверьте логи**:
   ```bash
   flyctl logs
   ```

2. **Проверьте статус**:
   ```bash
   flyctl status
   ```

3. **Проверьте секреты**:
   ```bash
   flyctl secrets list
   ```

### Проблемы с webhook

1. **Проверьте URL webhook**:
   ```bash
   curl https://tgnip.fly.dev/healthz
   ```

2. **Проверьте логи бота**:
   ```bash
   flyctl logs | grep webhook
   ```

3. **Переустановите webhook**:
   ```bash
   ./scripts/setup_webhook.sh
   ```

4. **Проверьте информацию о webhook**:
   ```bash
   curl https://api.telegram.org/bot<TOKEN>/getWebhookInfo
   ```

### Проблемы с Telegram API

1. **Проверьте токен бота**:
   ```bash
   flyctl secrets list | grep TELEGRAM_BOT_TOKEN
   ```

2. **Проверьте логи**:
   ```bash
   flyctl logs | grep "Bot"
   ```

## Конфигурация

### fly.toml

Основные настройки в `fly.toml`:

```toml
app = 'tgnip'
primary_region = 'waw'

[env]
  PORT = '8080'
  WEBHOOK_PORT = '8080'

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = 'stop'
  auto_start_machines = true
  min_machines_running = 0

[[services.http_checks]]
  interval = '15s'
  timeout = '2s'
  grace_period = '10s'
  method = 'GET'
  path = '/healthz'
```

### Регионы

Для изменения региона:

```bash
flyctl regions set waw
```

Доступные регионы: `waw`, `iad`, `lhr`, `hkg`, `syd`, `gru`, `nrt`

## Стоимость

- **Shared CPU**: $1.94/месяц за машину
- **Dedicated CPU**: $7.50/месяц за машину
- **Memory**: $0.50/GB/месяц

Для экономии используйте `auto_stop_machines = 'stop'` в конфигурации.
