package main

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

	return config
}
