package internal

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// WebhookServer представляет webhook сервер
type WebhookServer struct {
	bot         *tgbotapi.BotAPI
	config      *Config
	server      *http.Server
	webhookURL  string
	secretToken string
}

// NewWebhookServer создает новый webhook сервер
func NewWebhookServer(bot *tgbotapi.BotAPI, config *Config) *WebhookServer {
	// Получаем секретный токен из переменной окружения
	secretToken := os.Getenv("WEBHOOK_SECRET_TOKEN")
	if secretToken == "" {
		logger.Fatal("WEBHOOK_SECRET_TOKEN не установлен. Установите секретный токен в переменной окружения")
	}

	logger.Info("Используем WEBHOOK_SECRET_TOKEN из переменной окружения")

	return &WebhookServer{
		bot:         bot,
		config:      config,
		secretToken: secretToken,
	}
}

// Start запускает webhook сервер
func (ws *WebhookServer) Start() error {
	// Получаем настройки из переменных окружения
	port := os.Getenv("WEBHOOK_PORT")
	if port == "" {
		port = "8080"
	}

	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("WEBHOOK_URL не установлен")
	}

	ws.webhookURL = webhookURL

	// Настраиваем маршруты
	mux := http.NewServeMux()
	mux.HandleFunc("/telegram/webhook", ws.handleWebhook)
	mux.HandleFunc("/healthz", ws.handleHealth)

	// Создаем сервер
	ws.server = &http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: mux,
	}

	// Настраиваем TLS
	certFile := os.Getenv("SSL_CERT_FILE")
	keyFile := os.Getenv("SSL_KEY_FILE")

	if certFile != "" && keyFile != "" {
		// HTTPS режим
		logger.Infof("Запуск HTTPS webhook сервера на 0.0.0.0:%s", port)
		return ws.startHTTPS(certFile, keyFile)
	} else {
		// HTTP режим (только для разработки)
		logger.Warn("SSL сертификаты не найдены, запуск в HTTP режиме (только для разработки)")
		logger.Infof("Запуск HTTP webhook сервера на 0.0.0.0:%s", port)
		return ws.startHTTP()
	}
}

// startHTTPS запускает HTTPS сервер
func (ws *WebhookServer) startHTTPS(certFile, keyFile string) error {
	// Настраиваем TLS
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	ws.server.TLSConfig = tlsConfig

	// Устанавливаем webhook только для HTTPS
	if strings.HasPrefix(ws.webhookURL, "https://") {
		if err := ws.SetWebhook(); err != nil {
			return fmt.Errorf("ошибка установки webhook: %w", err)
		}
	}

	// Запускаем сервер в горутине
	go func() {
		if err := ws.server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			logger.Errorf("Ошибка HTTPS сервера: %v", err)
		}
	}()

	logger.Info("Webhook сервер запущен успешно")
	return nil
}

// startHTTP запускает HTTP сервер (только для разработки)
func (ws *WebhookServer) startHTTP() error {
	// Для разработки можно использовать ngrok или аналогичные сервисы
	// которые предоставляют HTTPS туннель к локальному HTTP серверу
	logger.Info("Для разработки используйте ngrok: ngrok http 8080")
	logger.Info("Затем установите WEBHOOK_URL=https://your-ngrok-url.ngrok.io")

	// Запускаем сервер в горутине
	go func() {
		if err := ws.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("Ошибка HTTP сервера: %v", err)
		}
	}()

	logger.Info("HTTP сервер запущен на 0.0.0.0:8080 (ожидание HTTPS туннеля)")
	return nil
}

// Stop останавливает webhook сервер
func (ws *WebhookServer) Stop() error {
	// Удаляем webhook только если бот доступен
	if ws.bot != nil {
		if err := ws.deleteWebhook(); err != nil {
			logger.Errorf("Ошибка удаления webhook: %v", err)
		}
	}

	// Останавливаем сервер
	if ws.server != nil {
		return ws.server.Close()
	}
	return nil
}

// setWebhook устанавливает webhook в Telegram
func (ws *WebhookServer) SetWebhook() error {
	webhookURL := ws.webhookURL + "/telegram/webhook"

	// Создаем webhook конфигурацию
	webhook, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		return fmt.Errorf("ошибка создания webhook: %w", err)
	}

	// Примечание: секретный токен проверяется на стороне сервера
	// при получении webhook запросов, но не передается в Telegram API
	logger.Info("Секретный токен настроен для проверки входящих запросов")

	// Устанавливаем webhook через API
	_, err = ws.bot.Request(webhook)
	if err != nil {
		return fmt.Errorf("не удалось установить webhook: %w", err)
	}

	logger.Infof("Webhook установлен: %s", webhookURL)
	return nil
}

// deleteWebhook удаляет webhook
func (ws *WebhookServer) deleteWebhook() error {
	_, err := ws.bot.Request(tgbotapi.DeleteWebhookConfig{})
	if err != nil {
		return fmt.Errorf("не удалось удалить webhook: %w", err)
	}

	logger.Info("Webhook удален")
	return nil
}

// handleWebhook обрабатывает входящие webhook запросы
func (ws *WebhookServer) handleWebhook(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем Content-Type
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "Invalid content type", http.StatusBadRequest)
		return
	}

	// Проверяем секретный токен (если установлен)
	if ws.secretToken != "" {
		token := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
		if token != ws.secretToken {
			logger.Warn("Неверный секретный токен webhook")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Читаем тело запроса
	var update tgbotapi.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		logger.Errorf("Ошибка декодирования webhook: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Проверяем, что это сообщение
	if update.Message == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Логируем входящее сообщение
	logger.Infof("Получено webhook сообщение от пользователя %d", update.Message.Chat.ID)

	// Обрабатываем сообщение напрямую только если бот доступен
	if ws.bot != nil {
		HandleMessage(ws.bot, update.Message)
	} else {
		logger.Debug("Bot не доступен, пропускаем обработку сообщения")
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleHealth обрабатывает health check запросы
func (ws *WebhookServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":  "healthy",
		"webhook": ws.webhookURL + "/telegram/webhook",
	}

	json.NewEncoder(w).Encode(response)
}
