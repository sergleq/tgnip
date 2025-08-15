# Интеграция gocolly/colly для извлечения контента

## Обзор

Проект теперь использует [gocolly/colly](https://github.com/gocolly/colly) для более качественного извлечения данных перед обработкой go-readability. Colly предоставляет мощные возможности для веб-скрапинга, включая:

- Обработка JavaScript-рендеринга
- Управление сессиями и cookies
- Ротация User-Agent
- Обработка редиректов
- Ограничение скорости запросов
- Обработка ошибок и повторные попытки

## Основные функции

### ExtractContent(url string) (*Content, error)
Основная функция для извлечения контента с настройками Colly по умолчанию.

```go
content, err := internal.ExtractContent("https://example.com/article")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Заголовок: %s\n", content.Title)
```

### ExtractContentWithConfig(url string, config *CollyConfig) (*Content, error)
Извлечение контента с пользовательской конфигурацией Colly.

```go
config := &internal.CollyConfig{
    UserAgent:      "Custom User Agent",
    Timeout:        45 * time.Second,
    MaxRetries:     3,
    FollowRedirect: true,
    RespectRobots:  false,
}

content, err := internal.ExtractContentWithConfig("https://example.com", config)
```

### ExtractContentWithFallback(url string) (*Content, error)
Извлечение контента с автоматическим fallback на стандартный HTTP клиент.

```go
content, err := internal.ExtractContentWithFallback("https://example.com")
```

## Конфигурация Colly

### CollyConfig

```go
type CollyConfig struct {
    UserAgent      string        // User-Agent для запросов
    Timeout        time.Duration // Таймаут запросов
    MaxRetries     int           // Максимальное количество повторных попыток
    FollowRedirect bool          // Следовать редиректам
    RespectRobots  bool          // Соблюдать robots.txt
}
```

### Настройки по умолчанию

```go
func DefaultCollyConfig() *CollyConfig {
    return &CollyConfig{
        UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
        Timeout:        30 * time.Second,
        MaxRetries:     3,
        FollowRedirect: true,
        RespectRobots:  false,
    }
}
```

## Возможности Colly

### 1. Ротация User-Agent
Автоматическая смена User-Agent для избежания блокировки:

```go
extensions.RandomUserAgent(c)
```

### 2. Ограничение скорости
Настройка задержек между запросами:

```go
c.Limit(&colly.LimitRule{
    DomainGlob:  "*",
    Parallelism: 1,
    RandomDelay: 1 * time.Second,
    Delay:       2 * time.Second,
})
```

### 3. Обработка ошибок
Автоматическая обработка ошибок сети:

```go
c.OnError(func(r *colly.Response, err error) {
    fmt.Printf("Ошибка при загрузке %s: %v\n", r.Request.URL, err)
})
```

### 4. Расширения
Использование встроенных расширений:

```go
extensions.RandomUserAgent(c)
extensions.Referer(c)
```

## Преимущества перед стандартным HTTP клиентом

1. **Лучшая обработка JavaScript**: Colly может обрабатывать динамический контент
2. **Управление сессиями**: Сохранение cookies между запросами
3. **Ротация User-Agent**: Автоматическая смена заголовков
4. **Ограничение скорости**: Встроенные механизмы для соблюдения правил сайтов
5. **Обработка ошибок**: Более надежная обработка сетевых ошибок
6. **Редиректы**: Автоматическая обработка HTTP редиректов

## Fallback механизм

Если Colly не может загрузить страницу, система автоматически переключается на стандартный HTTP клиент:

```go
// Сначала пробуем Colly
content, err := ExtractContent(url)
if err != nil {
    // Fallback на стандартный HTTP клиент
    return extractContentWithHTTPClient(url)
}
```

## Примеры использования

### Базовое использование

```go
package main

import (
    "fmt"
    "tgnip/internal"
)

func main() {
    content, err := internal.ExtractContent("https://habr.com/ru/articles/123456/")
    if err != nil {
        fmt.Printf("Ошибка: %v\n", err)
        return
    }
    
    fmt.Printf("Заголовок: %s\n", content.Title)
    fmt.Printf("Автор: %s\n", content.Author)
    fmt.Printf("Дата: %s\n", content.Date)
    fmt.Printf("Контент: %s\n", content.Markdown[:200])
}
```

### Пользовательская конфигурация

```go
config := &internal.CollyConfig{
    UserAgent:      "MyBot/1.0",
    Timeout:        60 * time.Second,
    MaxRetries:     5,
    FollowRedirect: true,
    RespectRobots:  true,
}

content, err := internal.ExtractContentWithConfig(url, config)
```

### Массовое извлечение

```go
urls := []string{
    "https://example1.com",
    "https://example2.com",
    "https://example3.com",
}

for _, url := range urls {
    go func(u string) {
        content, err := internal.ExtractContentWithFallback(u)
        if err != nil {
            fmt.Printf("Ошибка для %s: %v\n", u, err)
            return
        }
        // Обработка контента
        processContent(content)
    }(url)
}
```

## Тестирование

Запустите пример для тестирования интеграции:

```bash
go run examples/colly_example.go
```

## Зависимости

Добавленные зависимости:
- `github.com/gocolly/colly/v2` - основной пакет Colly
- `github.com/gocolly/colly/v2/extensions` - расширения для Colly

## Совместимость

- Go 1.23+ (требуется для Colly v2)
- Совместимо с существующим API
- Обратная совместимость с предыдущими версиями

## Производительность

Colly обеспечивает:
- Более высокую скорость извлечения данных
- Лучшую надежность при работе с различными сайтами
- Автоматическую обработку ошибок
- Оптимизированное использование ресурсов
