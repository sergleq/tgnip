package main

import (
	"os"
	"strconv"
)

// Config представляет конфигурацию приложения
type Config struct {
	// Параллелизм
	WorkerCount     int
	QueueBufferSize int
	RateLimitPerSec int

	// HTTP клиент
	HTTPTimeout int // в секундах
	MaxRetries  int

	// Логирование
	LogLevel        string
	MetricsInterval int // в минутах
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() *Config {
	config := &Config{
		// Значения по умолчанию
		WorkerCount:     5,
		QueueBufferSize: 10,
		RateLimitPerSec: 20,
		HTTPTimeout:     30,
		MaxRetries:      3,
		LogLevel:        "info",
		MetricsInterval: 5,
	}

	// Загружаем из переменных окружения
	if val := os.Getenv("WORKER_COUNT"); val != "" {
		if count, err := strconv.Atoi(val); err == nil && count > 0 {
			config.WorkerCount = count
		}
	}

	if val := os.Getenv("QUEUE_BUFFER_SIZE"); val != "" {
		if size, err := strconv.Atoi(val); err == nil && size > 0 {
			config.QueueBufferSize = size
		}
	}

	if val := os.Getenv("RATE_LIMIT_PER_SEC"); val != "" {
		if rate, err := strconv.Atoi(val); err == nil && rate > 0 {
			config.RateLimitPerSec = rate
		}
	}

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

	if val := os.Getenv("METRICS_INTERVAL"); val != "" {
		if interval, err := strconv.Atoi(val); err == nil && interval > 0 {
			config.MetricsInterval = interval
		}
	}

	return config
}

// GetRateLimitInterval возвращает интервал для rate limiting в миллисекундах
func (c *Config) GetRateLimitInterval() int {
	if c.RateLimitPerSec <= 0 {
		return 50 // 20 запросов в секунду по умолчанию
	}
	return 1000 / c.RateLimitPerSec
}
