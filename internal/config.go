package internal

import (
	"os"
	"strconv"
)

// Config представляет конфигурацию приложения
type Config struct {
	// HTTP клиент
	HTTPTimeout int // в секундах
	MaxRetries  int

	// Webhook
	WebhookURL  string
	WebhookPort string
	SSLCertFile string
	SSLKeyFile  string

	// Логирование
	LogLevel string

	// Ethical scraping
	UserAgent         string
	ContactEmail      string
	WhitelistDomains  string
	RateLimitInterval int // в секундах
	CacheTTL          int // в часах
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() *Config {
	config := &Config{
		// Значения по умолчанию
		HTTPTimeout: 30,
		MaxRetries:  3,
		LogLevel:    "info",
		WebhookPort: "8443",
	}

	// Загружаем из переменных окружения
	if val := os.Getenv("HTTP_TIMEOUT"); val != "" {
		if timeout, err := strconv.Atoi(val); err == nil && timeout > 0 {
			config.HTTPTimeout = timeout
		}
	}

	if val := os.Getenv("MAX_RETRIES"); val != "" {
		if retries, err := strconv.Atoi(val); err == nil && retries >= 0 {
			config.MaxRetries = retries
		}
	}

	if val := os.Getenv("LOG_LEVEL"); val != "" {
		config.LogLevel = val
	}

	// Webhook настройки
	if val := os.Getenv("WEBHOOK_URL"); val != "" {
		config.WebhookURL = val
	}

	if val := os.Getenv("WEBHOOK_PORT"); val != "" {
		config.WebhookPort = val
	}

	if val := os.Getenv("SSL_CERT_FILE"); val != "" {
		config.SSLCertFile = val
	}

	if val := os.Getenv("SSL_KEY_FILE"); val != "" {
		config.SSLKeyFile = val
	}

	// Ethical scraping настройки
	if val := os.Getenv("USER_AGENT"); val != "" {
		config.UserAgent = val
	} else {
		config.UserAgent = "TGNIP-Bot/1.0 (+https://github.com/sergleq/tgnip)"
	}

	if val := os.Getenv("CONTACT_EMAIL"); val != "" {
		config.ContactEmail = val
	}

	if val := os.Getenv("WHITELIST_DOMAINS"); val != "" {
		config.WhitelistDomains = val
	}

	if val := os.Getenv("RATE_LIMIT_INTERVAL"); val != "" {
		if interval, err := strconv.Atoi(val); err == nil && interval > 0 {
			config.RateLimitInterval = interval
		} else {
			config.RateLimitInterval = 1 // по умолчанию 1 секунда
		}
	} else {
		config.RateLimitInterval = 1
	}

	if val := os.Getenv("CACHE_TTL"); val != "" {
		if ttl, err := strconv.Atoi(val); err == nil && ttl > 0 {
			config.CacheTTL = ttl
		} else {
			config.CacheTTL = 24 // по умолчанию 24 часа
		}
	} else {
		config.CacheTTL = 24
	}

	return config
}
