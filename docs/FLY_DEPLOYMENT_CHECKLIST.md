# ✅ Чек-лист деплоя на Fly.io

## Предварительная подготовка

- [ ] Установлен `flyctl`
- [ ] Выполнена авторизация: `flyctl auth login`
- [ ] Создано приложение: `flyctl apps create tgnip`

## Конфигурация

- [ ] `fly.toml` настроен корректно
- [ ] `Dockerfile` оптимизирован для fly.io
- [ ] `.dockerignore` исключает ненужные файлы
- [ ] Health check эндпойнт `/healthz` работает

## Секреты

- [ ] `TELEGRAM_BOT_TOKEN` установлен
- [ ] `WEBHOOK_URL` установлен (например, https://tgnip.fly.dev)
- [ ] `WEBHOOK_SECRET_TOKEN` установлен (для безопасности webhook)
- [ ] Опциональные секреты настроены (LOG_LEVEL, HTTP_TIMEOUT, MAX_RETRIES)

## Деплой

- [ ] Приложение собирается без ошибок
- [ ] Деплой проходит успешно: `flyctl deploy`
- [ ] Health check проходит: `curl https://tgnip.fly.dev/healthz`
- [ ] Логи показывают успешный запуск
- [ ] Webhook установлен вручную: `./scripts/setup_webhook.sh`

## Мониторинг

- [ ] Статус приложения: `flyctl status`
- [ ] Логи приложения: `flyctl logs`
- [ ] Метрики: `flyctl dashboard`

## Тестирование

- [ ] Webhook эндпойнт отвечает: `curl -X POST https://tgnip.fly.dev/telegram/webhook`
- [ ] Telegram webhook установлен корректно: `curl https://api.telegram.org/bot<TOKEN>/getWebhookInfo`
- [ ] Бот отвечает на сообщения
- [ ] Секретный токен работает корректно

## Автоматизация (опционально)

- [ ] GitHub Actions workflow настроен
- [ ] `FLY_API_TOKEN` добавлен в GitHub Secrets
- [ ] Автоматический деплой работает при push в main

## Команды для проверки

```bash
# Проверка статуса
flyctl status

# Проверка секретов
flyctl secrets list

# Проверка health check
curl https://tgnip.fly.dev/healthz

# Просмотр логов
flyctl logs

# Мониторинг
flyctl dashboard
```

## Troubleshooting

### Если деплой не проходит:
1. Проверьте логи: `flyctl logs`
2. Проверьте статус: `flyctl status`
3. Проверьте секреты: `flyctl secrets list`
4. Перезапустите: `flyctl restart`

### Если webhook не работает:
1. Проверьте WEBHOOK_URL в секретах
2. Проверьте health check
3. Проверьте логи на ошибки Telegram API
4. Переустановите webhook: `flyctl restart`

### Если бот не отвечает:
1. Проверьте TELEGRAM_BOT_TOKEN
2. Проверьте логи на ошибки
3. Проверьте, что webhook установлен в Telegram
