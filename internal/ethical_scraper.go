package internal

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
)

// EthicalScraper представляет этичный веб-скрапер
type EthicalScraper struct {
	client    *http.Client
	userAgent string
	contact   string
	rateLimit *RateLimiter
	cache     *ResponseCache
	whitelist *DomainWhitelist
}

// RateLimiter управляет частотой запросов
type RateLimiter struct {
	requests map[string]time.Time
	interval time.Duration
}

// ResponseCache кэширует ответы
type ResponseCache struct {
	cache map[string]*CachedResponse
	ttl   time.Duration
}

// CachedResponse представляет кэшированный ответ
type CachedResponse struct {
	Content      []byte
	Headers      http.Header
	StatusCode   int
	ExpiresAt    time.Time
	ETag         string
	LastModified string
}

// DomainWhitelist управляет белым списком доменов
type DomainWhitelist struct {
	allowed map[string]bool
	apiKeys map[string]string
}

// ScrapingResult представляет результат скрапинга
type ScrapingResult struct {
	Content     []byte
	StatusCode  int
	Headers     http.Header
	IsBlocked   bool
	BlockReason string
	IsCached    bool
}

// NewEthicalScraper создает новый этичный скрапер
func NewEthicalScraper(config *Config) *EthicalScraper {
	// Создаем HTTP клиент с правильными настройками
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	// Поддержка HTTP/2
	if err := http2.ConfigureTransport(transport); err != nil {
		logrus.Warnf("HTTP/2 не поддерживается: %v", err)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(config.HTTPTimeout) * time.Second,
	}

	// Настройка User-Agent и контактной информации
	userAgent := config.UserAgent
	contact := config.ContactEmail

	// Инициализируем белый список
	whitelist := &DomainWhitelist{
		allowed: make(map[string]bool),
		apiKeys: make(map[string]string),
	}

	// Загружаем домены из белого списка
	if config.WhitelistDomains != "" {
		domains := strings.Split(config.WhitelistDomains, ",")
		for _, domain := range domains {
			domain = strings.TrimSpace(domain)
			if domain != "" {
				whitelist.AddToWhitelist(domain)
			}
		}
	}

	return &EthicalScraper{
		client:    client,
		userAgent: userAgent,
		contact:   contact,
		rateLimit: &RateLimiter{
			requests: make(map[string]time.Time),
			interval: time.Duration(config.RateLimitInterval) * time.Second,
		},
		cache: &ResponseCache{
			cache: make(map[string]*CachedResponse),
			ttl:   time.Duration(config.CacheTTL) * time.Hour,
		},
		whitelist: whitelist,
	}
}

// ScrapeURL этично извлекает контент из URL
func (s *EthicalScraper) ScrapeURL(ctx context.Context, pageURL string) (*ScrapingResult, error) {
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		return nil, fmt.Errorf("неверный URL: %w", err)
	}

	domain := parsedURL.Hostname()

	// Проверяем белый список
	if !s.whitelist.IsAllowed(domain) {
		return &ScrapingResult{
			IsBlocked:   true,
			BlockReason: "Домен не в белом списке",
		}, nil
	}

	// Проверяем robots.txt
	if err := s.checkRobotsTxt(parsedURL); err != nil {
		return &ScrapingResult{
			IsBlocked:   true,
			BlockReason: fmt.Sprintf("robots.txt запрещает доступ: %v", err),
		}, nil
	}

	// Проверяем кэш
	if cached := s.cache.Get(pageURL); cached != nil {
		return &ScrapingResult{
			Content:    cached.Content,
			Headers:    cached.Headers,
			StatusCode: cached.StatusCode,
			IsCached:   true,
		}, nil
	}

	// Rate limiting
	if err := s.rateLimit.Wait(domain); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	}

	// Создаем запрос с правильными заголовками
	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Устанавливаем этичные заголовки
	s.setEthicalHeaders(req, parsedURL)

	// Выполняем запрос с retry логикой
	var resp *http.Response
	for attempt := 0; attempt < 3; attempt++ {
		resp, err = s.client.Do(req)
		if err != nil {
			if attempt < 2 {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			return nil, fmt.Errorf("ошибка запроса: %w", err)
		}
		break
	}
	defer resp.Body.Close()

	// Проверяем статус код
	if resp.StatusCode == 403 || resp.StatusCode == 451 {
		return &ScrapingResult{
			IsBlocked:   true,
			BlockReason: fmt.Sprintf("Доступ запрещен (статус %d)", resp.StatusCode),
		}, nil
	}

	if resp.StatusCode == 429 {
		return &ScrapingResult{
			IsBlocked:   true,
			BlockReason: "Слишком много запросов (429)",
		}, nil
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("неожиданный статус код: %d", resp.StatusCode)
	}

	// Читаем контент
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Кэшируем результат
	s.cache.Set(pageURL, &CachedResponse{
		Content:      content,
		Headers:      resp.Header,
		StatusCode:   resp.StatusCode,
		ExpiresAt:    time.Now().Add(s.cache.ttl),
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
	})

	return &ScrapingResult{
		Content:    content,
		Headers:    resp.Header,
		StatusCode: resp.StatusCode,
		IsCached:   false,
	}, nil
}

// setEthicalHeaders устанавливает этичные заголовки
func (s *EthicalScraper) setEthicalHeaders(req *http.Request, parsedURL *url.URL) {
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	// Добавляем контактную информацию
	req.Header.Set("From", s.contact)
	req.Header.Set("X-Requested-With", "TGNIP-Bot")

	// Условные запросы для экономии трафика
	if cached := s.cache.Get(req.URL.String()); cached != nil {
		if cached.ETag != "" {
			req.Header.Set("If-None-Match", cached.ETag)
		}
		if cached.LastModified != "" {
			req.Header.Set("If-Modified-Since", cached.LastModified)
		}
	}
}

// checkRobotsTxt проверяет robots.txt
func (s *EthicalScraper) checkRobotsTxt(parsedURL *url.URL) error {
	robotsURL := fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, parsedURL.Host)

	req, err := http.NewRequest("GET", robotsURL, nil)
	if err != nil {
		return fmt.Errorf("ошибка создания запроса robots.txt: %w", err)
	}

	req.Header.Set("User-Agent", s.userAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		// Если robots.txt недоступен, считаем что доступ разрешен
		logrus.Warnf("robots.txt недоступен для %s: %v", parsedURL.Host, err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		// robots.txt не найден, доступ разрешен
		return nil
	}

	if resp.StatusCode != 200 {
		// Если robots.txt недоступен, считаем что доступ разрешен
		logrus.Warnf("robots.txt недоступен для %s (статус %d)", parsedURL.Host, resp.StatusCode)
		return nil
	}

	// Парсим robots.txt
	scanner := bufio.NewScanner(resp.Body)
	currentUserAgent := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "User-agent:") {
			currentUserAgent = strings.TrimSpace(strings.TrimPrefix(line, "User-agent:"))
		} else if strings.HasPrefix(line, "Disallow:") && (currentUserAgent == "*" || currentUserAgent == "TGNIP-Bot") {
			disallowPath := strings.TrimSpace(strings.TrimPrefix(line, "Disallow:"))
			if strings.HasPrefix(parsedURL.Path, disallowPath) {
				return fmt.Errorf("путь %s запрещен в robots.txt", parsedURL.Path)
			}
		}
	}

	return nil
}

// Wait ожидает соблюдения rate limit
func (rl *RateLimiter) Wait(domain string) error {
	if lastRequest, exists := rl.requests[domain]; exists {
		timeSinceLast := time.Since(lastRequest)
		if timeSinceLast < rl.interval {
			sleepTime := rl.interval - timeSinceLast
			time.Sleep(sleepTime)
		}
	}

	rl.requests[domain] = time.Now()
	return nil
}

// Get получает кэшированный ответ
func (rc *ResponseCache) Get(key string) *CachedResponse {
	if cached, exists := rc.cache[key]; exists && time.Now().Before(cached.ExpiresAt) {
		return cached
	}

	// Удаляем устаревший кэш
	delete(rc.cache, key)
	return nil
}

// Set сохраняет ответ в кэш
func (rc *ResponseCache) Set(key string, response *CachedResponse) {
	rc.cache[key] = response
}

// IsAllowed проверяет, разрешен ли домен
func (dw *DomainWhitelist) IsAllowed(domain string) bool {
	// Проверяем точное совпадение
	if allowed, exists := dw.allowed[domain]; exists {
		return allowed
	}

	// Проверяем поддомены
	for allowedDomain, allowed := range dw.allowed {
		if strings.HasSuffix(domain, "."+allowedDomain) || domain == allowedDomain {
			return allowed
		}
	}

	// По умолчанию разрешаем все домены (серый список)
	return true
}

// AddToWhitelist добавляет домен в белый список
func (dw *DomainWhitelist) AddToWhitelist(domain string) {
	dw.allowed[domain] = true
}

// RemoveFromWhitelist удаляет домен из белого списка
func (dw *DomainWhitelist) RemoveFromWhitelist(domain string) {
	dw.allowed[domain] = false
}

// SetAPIKey устанавливает API ключ для домена
func (dw *DomainWhitelist) SetAPIKey(domain, apiKey string) {
	dw.apiKeys[domain] = apiKey
}

// GetAPIKey получает API ключ для домена
func (dw *DomainWhitelist) GetAPIKey(domain string) string {
	return dw.apiKeys[domain]
}
