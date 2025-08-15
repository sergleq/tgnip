# Webhook для Telegram Bot

## 📊 Обзор

Бот поддерживает два режима работы:
- **Polling** (по умолчанию) - для разработки и тестирования
- **Webhook** - для продакшена с высокой нагрузкой

## 🚀 Преимущества Webhook

### ✅ По сравнению с Polling:
- **Мгновенная реакция** - сообщения обрабатываются сразу
- **Эффективность** - нет постоянных HTTP запросов
- **Масштабируемость** - подходит для больших нагрузок
- **Низкое потребление ресурсов** - меньше CPU и памяти
- **Надежность** - меньше зависимость от сетевых проблем

### ❌ Требования:
- **Публичный HTTPS сервер**
- **SSL сертификат**
- **Открытый порт 443/80**
- **Стабильное интернет-соединение**

## ⚙️ Конфигурация

### Переменные окружения:

```bash
# Обязательные для webhook
WEBHOOK_URL=https://your-domain.com

# Опциональные
WEBHOOK_PORT=8443                    # Порт сервера (по умолчанию 8443)
SSL_CERT_FILE=/path/to/cert.pem      # SSL сертификат
SSL_KEY_FILE=/path/to/key.pem        # SSL ключ
```

### Автоматическое определение режима:

```go
// Если WEBHOOK_URL установлен - webhook режим
// Если WEBHOOK_URL не установлен - polling режим
webhookURL := os.Getenv("WEBHOOK_URL")
if webhookURL != "" {
    // Webhook режим
} else {
    // Polling режим
}
```

## 🔧 Техническая реализация

### WebhookServer структура:

```go
type WebhookServer struct {
    bot         *tgbotapi.BotAPI
    workerPool  *WorkerPool
    config      *Config
    server      *http.Server
    webhookURL  string
    secretToken string
}
```

### Основные компоненты:

1. **HTTP/HTTPS сервер** - принимает webhook запросы
2. **Webhook endpoint** - `/webhook` для обработки сообщений
3. **Health check** - `/health` для мониторинга
4. **Worker Pool** - параллельная обработка сообщений
5. **Security** - проверка секретного токена

## 🌐 Endpoints

### POST /webhook
Обрабатывает входящие сообщения от Telegram.

**Заголовки:**
```
Content-Type: application/json
X-Telegram-Bot-Api-Secret-Token: secret_1234567890
```

**Тело запроса:**
```json
{
  "update_id": 123456789,
  "message": {
    "message_id": 1,
    "from": {
      "id": 123456789,
      "first_name": "John",
      "language_code": "en"
    },
    "chat": {
      "id": 123456789,
      "type": "private"
    },
    "text": "https://example.com"
  }
}
```

**Ответ:**
```
HTTP/1.1 200 OK
OK
```

### GET /health
Health check endpoint для мониторинга.

**Ответ:**
```json
{
  "status": "healthy",
  "webhook": "https://your-domain.com/webhook",
  "workers": 5,
  "queueSize": 10
}
```

## 🔒 Безопасность

### Секретный токен:
- Автоматически генерируется при запуске
- Проверяется в заголовке `X-Telegram-Bot-Api-Secret-Token`
- Защищает от несанкционированных запросов

### SSL/TLS:
- Обязательно для продакшена
- Минимальная версия TLS 1.2
- Поддержка Let's Encrypt сертификатов

## 📈 Мониторинг

### Логи:
```
INFO[2025-08-13T14:30:00Z] Запуск в webhook режиме
INFO[2025-08-13T14:30:01Z] Webhook установлен: https://your-domain.com/webhook
INFO[2025-08-13T14:30:02Z] Webhook сервер запущен успешно
INFO[2025-08-13T14:30:15Z] Получено webhook сообщение от пользователя 123456789
```

### Health check:
```bash
curl https://your-domain.com/health
```

## 🚀 Развертывание

### 1. Подготовка сервера:
```bash
# Установка SSL сертификата
sudo certbot certonly --standalone -d your-domain.com

# Настройка переменных окружения
export WEBHOOK_URL=https://your-domain.com
export SSL_CERT_FILE=/etc/letsencrypt/live/your-domain.com/fullchain.pem
export SSL_KEY_FILE=/etc/letsencrypt/live/your-domain.com/privkey.pem
```

### 2. Запуск бота:
```bash
./tgnip
```

### 3. Проверка работы:
```bash
# Health check
curl https://your-domain.com/health

# Проверка webhook в Telegram
curl -X POST https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getWebhookInfo
```

## 🔄 Миграция с Polling

### Пошаговый план:

1. **Подготовка инфраструктуры:**
   - Настройка домена и SSL
   - Открытие портов (443, 80)

2. **Обновление конфигурации:**
   ```bash
   export WEBHOOK_URL=https://your-domain.com
   export SSL_CERT_FILE=/path/to/cert.pem
   export SSL_KEY_FILE=/path/to/key.pem
   ```

3. **Перезапуск бота:**
   ```bash
   pkill -f tgnip
   ./tgnip
   ```

4. **Проверка работы:**
   - Отправка тестового сообщения
   - Проверка логов
   - Мониторинг метрик

## 🛠️ Устранение неполадок

### Частые проблемы:

#### 1. Webhook не устанавливается:
```bash
# Проверьте:
- Доступность домена из интернета
- SSL сертификат
- Правильность WEBHOOK_URL
```

#### 2. Сообщения не приходят:
```bash
# Проверьте:
- Логи сервера
- Настройки firewall
- Статус webhook в Telegram API
```

#### 3. SSL ошибки:
```bash
# Проверьте:
- Валидность сертификата
- Права доступа к файлам
- Версию TLS
```

### Отладка:
```bash
# Включите debug логирование
export LOG_LEVEL=debug

# Проверьте статус webhook
curl -X POST https://api.telegram.org/bot<TOKEN>/getWebhookInfo

# Проверьте health endpoint
curl https://your-domain.com/health
```

## 📊 Сравнение режимов

| Характеристика | Polling | Webhook |
|----------------|---------|---------|
| Простота настройки | ✅ Легко | ❌ Сложнее |
| Требования к серверу | ❌ Нет | ✅ HTTPS, публичный IP |
| Задержка обработки | ❌ До 60 сек | ✅ Мгновенно |
| Потребление ресурсов | ❌ Высокое | ✅ Низкое |
| Масштабируемость | ❌ Ограничена | ✅ Высокая |
| Надежность | ❌ Средняя | ✅ Высокая |

## 🔮 Будущие улучшения

1. **Автоматическое обновление SSL** - интеграция с Let's Encrypt
2. **Load balancing** - поддержка нескольких экземпляров
3. **Rate limiting** - защита от DDoS
4. **Metrics dashboard** - веб-интерфейс для мониторинга
5. **Graceful shutdown** - корректное завершение при обновлении сертификатов
