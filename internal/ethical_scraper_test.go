package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewEthicalScraper(t *testing.T) {
	config := &Config{
		HTTPTimeout:       30,
		RateLimitInterval: 2,
		CacheTTL:          1,
		UserAgent:         "Test-Bot/1.0",
		ContactEmail:      "test@example.com",
		WhitelistDomains:  "example.com,test.com",
	}

	scraper := NewEthicalScraper(config)

	if scraper == nil {
		t.Fatal("NewEthicalScraper вернул nil")
	}

	if scraper.userAgent != "Test-Bot/1.0" {
		t.Errorf("Ожидался User-Agent 'Test-Bot/1.0', получен '%s'", scraper.userAgent)
	}

	if scraper.contact != "test@example.com" {
		t.Errorf("Ожидался контакт 'test@example.com', получен '%s'", scraper.contact)
	}

	if scraper.rateLimit.interval != 2*time.Second {
		t.Errorf("Ожидался интервал rate limit 2 секунды, получен %v", scraper.rateLimit.interval)
	}

	if scraper.cache.ttl != 1*time.Hour {
		t.Errorf("Ожидался TTL кэша 1 час, получен %v", scraper.cache.ttl)
	}
}

func TestDomainWhitelist(t *testing.T) {
	whitelist := &DomainWhitelist{
		allowed: make(map[string]bool),
		apiKeys: make(map[string]string),
	}

	// Тест добавления в белый список
	whitelist.AddToWhitelist("example.com")
	if !whitelist.IsAllowed("example.com") {
		t.Error("example.com должен быть разрешен")
	}

	// Тест поддоменов
	whitelist.AddToWhitelist("blog.example.com")
	if !whitelist.IsAllowed("blog.example.com") {
		t.Error("blog.example.com должен быть разрешен")
	}

	// Тест удаления из белого списка
	whitelist.RemoveFromWhitelist("example.com")
	if whitelist.IsAllowed("example.com") {
		t.Error("example.com должен быть запрещен после удаления")
	}

	// Тест API ключей
	whitelist.SetAPIKey("api.example.com", "test-key")
	if key := whitelist.GetAPIKey("api.example.com"); key != "test-key" {
		t.Errorf("Ожидался API ключ 'test-key', получен '%s'", key)
	}
}

func TestRateLimiter(t *testing.T) {
	rateLimiter := &RateLimiter{
		requests: make(map[string]time.Time),
		interval: 100 * time.Millisecond,
	}

	domain := "test.com"

	// Первый запрос должен пройти сразу
	start := time.Now()
	err := rateLimiter.Wait(domain)
	if err != nil {
		t.Errorf("Первый запрос не должен возвращать ошибку: %v", err)
	}

	// Второй запрос должен ждать
	err = rateLimiter.Wait(domain)
	if err != nil {
		t.Errorf("Второй запрос не должен возвращать ошибку: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed < 100*time.Millisecond {
		t.Errorf("Второй запрос должен был ждать минимум 100ms, но ждал только %v", elapsed)
	}
}

func TestResponseCache(t *testing.T) {
	cache := &ResponseCache{
		cache: make(map[string]*CachedResponse),
		ttl:   1 * time.Hour,
	}

	key := "test-url"
	response := &CachedResponse{
		Content:      []byte("test content"),
		StatusCode:   200,
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		ETag:         "test-etag",
		LastModified: "test-last-modified",
	}

	// Тест сохранения
	cache.Set(key, response)
	cached := cache.Get(key)
	if cached == nil {
		t.Fatal("Кэшированный ответ не найден")
	}

	if string(cached.Content) != "test content" {
		t.Errorf("Ожидался контент 'test content', получен '%s'", string(cached.Content))
	}

	// Тест устаревшего кэша
	expiredResponse := &CachedResponse{
		Content:   []byte("expired content"),
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Уже истек
	}
	cache.Set(key, expiredResponse)
	cached = cache.Get(key)
	if cached != nil {
		t.Error("Устаревший кэш должен быть удален")
	}
}

func TestEthicalScraperWithMockServer(t *testing.T) {
	// Создаем mock сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем заголовки только для основного запроса, не для robots.txt
		if r.URL.Path != "/robots.txt" {
			// Проверяем User-Agent
			if r.Header.Get("User-Agent") != "Test-Bot/1.0" {
				t.Errorf("Ожидался User-Agent 'Test-Bot/1.0', получен '%s'", r.Header.Get("User-Agent"))
			}

			// Проверяем контактную информацию
			if r.Header.Get("X-Requested-With") != "TGNIP-Bot" {
				t.Errorf("Ожидался X-Requested-With 'TGNIP-Bot', получен '%s'", r.Header.Get("X-Requested-With"))
			}
		}

		// Обрабатываем robots.txt
		if r.URL.Path == "/robots.txt" {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("User-agent: *\nAllow: /\n"))
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("ETag", "test-etag")
		w.Header().Set("Last-Modified", "test-last-modified")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>Test Content</h1></body></html>"))
	}))
	defer server.Close()

	config := &Config{
		HTTPTimeout:       30,
		RateLimitInterval: 1,
		CacheTTL:          1,
		UserAgent:         "Test-Bot/1.0",
		ContactEmail:      "test@example.com",
	}

	scraper := NewEthicalScraper(config)
	ctx := context.Background()

	// Тест успешного запроса
	result, err := scraper.ScrapeURL(ctx, server.URL)
	if err != nil {
		t.Fatalf("Ошибка при запросе: %v", err)
	}

	if result.IsBlocked {
		t.Error("Запрос не должен быть заблокирован")
	}

	if result.StatusCode != 200 {
		t.Errorf("Ожидался статус 200, получен %d", result.StatusCode)
	}

	// Проверяем, что результат не кэширован (первый запрос)
	if result.IsCached {
		t.Error("Первый запрос не должен быть кэширован")
	}

	// Тест кэшированного запроса
	result2, err := scraper.ScrapeURL(ctx, server.URL)
	if err != nil {
		t.Fatalf("Ошибка при повторном запросе: %v", err)
	}

	if !result2.IsCached {
		t.Error("Второй запрос должен использовать кэш")
	}
}

func TestRobotsTxtBlocking(t *testing.T) {
	// Создаем mock сервер с robots.txt
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("User-agent: TGNIP-Bot\nDisallow: /\n"))
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>Test Content</h1></body></html>"))
	}))
	defer server.Close()

	config := &Config{
		HTTPTimeout:       30,
		RateLimitInterval: 1,
		CacheTTL:          1,
		UserAgent:         "TGNIP-Bot",
		ContactEmail:      "test@example.com",
	}

	scraper := NewEthicalScraper(config)
	ctx := context.Background()

	// Тест блокировки robots.txt
	result, err := scraper.ScrapeURL(ctx, server.URL+"/test")
	if err != nil {
		t.Fatalf("Ошибка при запросе: %v", err)
	}

	if !result.IsBlocked {
		t.Error("Запрос должен быть заблокирован robots.txt")
	}

	if result.BlockReason == "" {
		t.Error("Должна быть указана причина блокировки")
	}
}

func TestRateLimitBlocking(t *testing.T) {
	// Создаем mock сервер, который возвращает 429
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Rate limit exceeded"))
	}))
	defer server.Close()

	config := &Config{
		HTTPTimeout:       30,
		RateLimitInterval: 1,
		CacheTTL:          1,
		UserAgent:         "Test-Bot/1.0",
		ContactEmail:      "test@example.com",
	}

	scraper := NewEthicalScraper(config)
	ctx := context.Background()

	// Тест блокировки rate limit
	result, err := scraper.ScrapeURL(ctx, server.URL)
	if err != nil {
		t.Fatalf("Ошибка при запросе: %v", err)
	}

	if !result.IsBlocked {
		t.Error("Запрос должен быть заблокирован из-за rate limit")
	}

	if result.BlockReason == "" {
		t.Error("Должна быть указана причина блокировки")
	}
}

func TestGeoBlocking(t *testing.T) {
	// Создаем mock сервер, который возвращает 403
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Access forbidden"))
	}))
	defer server.Close()

	config := &Config{
		HTTPTimeout:       30,
		RateLimitInterval: 1,
		CacheTTL:          1,
		UserAgent:         "Test-Bot/1.0",
		ContactEmail:      "test@example.com",
	}

	scraper := NewEthicalScraper(config)
	ctx := context.Background()

	// Тест блокировки geo-blocking
	result, err := scraper.ScrapeURL(ctx, server.URL)
	if err != nil {
		t.Fatalf("Ошибка при запросе: %v", err)
	}

	if !result.IsBlocked {
		t.Error("Запрос должен быть заблокирован из-за geo-blocking")
	}

	if result.BlockReason == "" {
		t.Error("Должна быть указана причина блокировки")
	}
}
