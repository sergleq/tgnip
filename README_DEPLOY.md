# 🚀 Быстрый деплой на Fly.io

## Предварительная настройка

1. **Установите flyctl**:
   ```bash
   curl -L https://fly.io/install.sh | sh
   ```

2. **Авторизуйтесь**:
   ```bash
   flyctl auth login
   ```

3. **Создайте приложение** (если нужно):
   ```bash
   flyctl apps create tgnip
   ```

## Деплой

### Автоматический деплой

1. **Настройте секреты**:
   ```bash
   ./scripts/setup_fly_secrets.sh
   ```

2. **Запустите деплой**:
   ```bash
   ./scripts/deploy.sh
   ```

3. **Установите webhook**:
   ```bash
   ./scripts/setup_webhook.sh
   ```

### Ручной деплой

1. **Установите секреты**:
   ```bash
   flyctl secrets set TELEGRAM_BOT_TOKEN="your_bot_token"
   flyctl secrets set WEBHOOK_URL="https://tgnip.fly.dev"
   flyctl secrets set WEBHOOK_SECRET_TOKEN="your_secret_token"
   ```

2. **Деплой**:
   ```bash
   flyctl deploy
   ```

3. **Установите webhook**:
   ```bash
   ./scripts/setup_webhook.sh
   ```

## Проверка

```bash
# Статус
flyctl status

# Health check
curl https://tgnip.fly.dev/healthz

# Логи
flyctl logs
```

## Полезные команды

```bash
# Перезапуск
flyctl restart

# Масштабирование
flyctl scale count 1

# Мониторинг
flyctl dashboard
```

## Документация

Подробная документация: [DEPLOY.md](DEPLOY.md)
