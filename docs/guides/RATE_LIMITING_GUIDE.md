# Руководство по работе с ограничениями сайтов

## 🚫 Типы ограничений

Многие сайты используют различные методы для ограничения доступа к своему контенту:

### 1. **Rate Limiting (Ограничение частоты запросов)**
- **HTTP 429 (Too Many Requests)** - превышен лимит запросов
- **Retry-After заголовок** - время ожидания перед повторной попыткой
- **IP-based ограничения** - блокировка по IP адресу
- **User-based ограничения** - ограничения для конкретных пользователей

### 2. **User-Agent блокировка**
- Блокировка ботов и скриптов
- Требование определенного User-Agent
- Проверка на "человеческое" поведение

### 3. **CAPTCHA и JavaScript**
- Требование выполнения CAPTCHA
- JavaScript-зависимая загрузка контента
- Cloudflare и подобные защиты

### 4. **Географические ограничения**
- Блокировка по IP адресу
- Региональные ограничения доступа

## 🛠 Реализованные решения

### ✅ **Retry логика с экспоненциальной задержкой**

```go
func getRetryDelay(attempt int) time.Duration {
    // Экспоненциальная задержка: 1s, 2s, 4s, 8s...
    delay := time.Duration(1<<uint(attempt)) * time.Second
    
    // Ограничиваем максимальную задержку 30 секундами
    if delay > 30*time.Second {
        delay = 30 * time.Second
    }
    
    return delay
}
```

**Преимущества:**
- Автоматическое увеличение задержки при ошибках
- Ограничение максимальной задержки
- Соблюдение Retry-After заголовков

### ✅ **Имитация браузера**

```go
func createRequest(url string) (*http.Request, error) {
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }

    // Устанавливаем User-Agent как у обычного браузера
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
    
    // Добавляем другие заголовки для имитации браузера
    req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
    req.Header.Set("Accept-Language", "en-US,en;q=0.5")
    req.Header.Set("Accept-Encoding", "gzip, deflate")
    req.Header.Set("Connection", "keep-alive")
    req.Header.Set("Upgrade-Insecure-Requests", "1")
    req.Header.Set("Sec-Fetch-Dest", "document")
    req.Header.Set("Sec-Fetch-Mode", "navigate")
    req.Header.Set("Sec-Fetch-Site", "none")
    req.Header.Set("Cache-Control", "max-age=0")

    return req, nil
}
```

**Преимущества:**
- Обход простых User-Agent фильтров
- Имитация реального браузера
- Поддержка современных заголовков безопасности

### ✅ **Умная обработка ошибок**

```go
func shouldRetry(statusCode int, err error) bool {
    if err != nil {
        return true // Повторяем при любых ошибках сети
    }
    
    // Повторяем при временных ошибках сервера
    switch statusCode {
    case 429: // Too Many Requests
        return true
    case 500, 502, 503, 504: // Server errors
        return true
    case 408: // Request Timeout
        return true
    default:
        return false
    }
}
```

**Преимущества:**
- Повторение только при временных ошибках
- Не повторяем при 404, 403 (постоянные ошибки)
- Автоматическое определение типа ошибки

## ⚙️ Конфигурация

### Переменные окружения

```env
# HTTP Client Configuration
HTTP_TIMEOUT=30                # Таймаут HTTP запросов в секундах
MAX_RETRIES=3                  # Максимальное количество повторных попыток
MAX_FILE_SIZE=10485760         # Максимальный размер файла в байтах (10MB)
REQUEST_DELAY=1000             # Задержка между запросами в миллисекундах
MAX_CONCURRENT_REQUESTS=5      # Максимальное количество одновременных запросов
```

### Настройки по умолчанию

| Параметр | Значение | Описание |
|----------|----------|----------|
| `HTTP_TIMEOUT` | 30 сек | Таймаут для HTTP запросов |
| `MAX_RETRIES` | 3 | Максимальное количество попыток |
| `MAX_FILE_SIZE` | 10MB | Максимальный размер загружаемого файла |
| `REQUEST_DELAY` | 1000 мс | Задержка между запросами |
| `MAX_CONCURRENT_REQUESTS` | 5 | Максимум одновременных запросов |

## 📊 Логирование

### Уровни логирования

```go
logger.Infof("🔍 Попытка %d/%d извлечения контента из: %s", attempt+1, config.MaxRetries+1, pageURL)
logger.Warnf("⚠️ Rate limit превышен для %s (попытка %d)", pageURL, attempt+1)
logger.Infof("⏳ Ожидание %v согласно Retry-After заголовку...", delay)
```

### Примеры логов

```
🔍 Попытка 1/4 извлечения контента из: https://example.com
⚠️ Rate limit превышен для https://example.com (попытка 1)
⏳ Ожидание 5s согласно Retry-After заголовку...
🔍 Попытка 2/4 извлечения контента из: https://example.com
✅ Контент успешно извлечен: заголовок='Example Article', автор='John Doe', дата='2024-01-01'
```

## 🚀 Рекомендации по использованию

### 1. **Настройка для разных сайтов**

```env
# Для агрессивных сайтов
HTTP_TIMEOUT=60
MAX_RETRIES=5
REQUEST_DELAY=2000

# Для стабильных сайтов
HTTP_TIMEOUT=15
MAX_RETRIES=2
REQUEST_DELAY=500
```

### 2. **Мониторинг ограничений**

```bash
# Проверка логов на rate limiting
grep "Rate limit" logs/app.log

# Проверка успешных запросов
grep "✅ Контент успешно извлечен" logs/app.log
```

### 3. **Обработка специфических сайтов**

Некоторые сайты могут требовать дополнительной настройки:

#### Cloudflare защищенные сайты
- Могут требовать JavaScript
- Часто используют CAPTCHA
- Могут блокировать автоматические запросы

#### Новостные сайты
- Часто имеют строгие rate limits
- Могут требовать подписку
- Используют географические ограничения

#### Социальные сети
- Обычно блокируют ботов
- Требуют аутентификацию
- Используют сложные алгоритмы защиты

## 🔧 Дополнительные улучшения

### 1. **Proxy поддержка**

```go
// Добавить поддержку прокси для обхода IP блокировок
transport := &http.Transport{
    Proxy: http.ProxyURL(proxyURL),
    // ... другие настройки
}
```

### 2. **Rotating User-Agents**

```go
var userAgents = []string{
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
}
```

### 3. **Session management**

```go
// Сохранение cookies между запросами
jar, _ := cookiejar.New(nil)
client := &http.Client{
    Jar: jar,
    // ... другие настройки
}
```

## ⚠️ Важные замечания

### 1. **Соблюдение robots.txt**
- Всегда проверяйте robots.txt перед запросами
- Уважайте правила сайта
- Не превышайте разумные лимиты

### 2. **Правовые аспекты**
- Убедитесь, что у вас есть право на извлечение контента
- Соблюдайте условия использования сайтов
- Не нарушайте авторские права

### 3. **Этичное использование**
- Не перегружайте серверы
- Используйте разумные задержки
- Уважайте ограничения сайтов

## 📈 Мониторинг производительности

### Метрики для отслеживания

```go
type Metrics struct {
    TotalRequests    int64
    SuccessfulRequests int64
    FailedRequests   int64
    RateLimitHits    int64
    AverageResponseTime time.Duration
}
```

### Алерты

- Увеличение количества 429 ошибок
- Снижение успешности запросов
- Увеличение времени ответа

## 🔄 Обновления и поддержка

### Регулярные обновления
- Обновление User-Agent строк
- Адаптация к новым методам защиты
- Улучшение retry логики

### Сообщество
- Отслеживание новых методов обхода ограничений
- Обмен опытом с другими разработчиками
- Участие в open source проектах
