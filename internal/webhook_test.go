package internal

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

func init() {
	// Инициализируем логгер для тестов
	logger = logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Только ошибки в тестах
}

func TestWebhookHealthEndpoint(t *testing.T) {
	// Создаем мок бота и конфигурации
	config := &Config{}

	// Создаем webhook сервер
	webhookServer := &WebhookServer{
		config:     config,
		webhookURL: "https://test.com",
	}

	// Создаем тестовый запрос
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Создаем ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(webhookServer.handleHealth)

	// Выполняем запрос
	handler.ServeHTTP(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Проверяем Content-Type
	expected := "application/json"
	if rr.Header().Get("Content-Type") != expected {
		t.Errorf("handler returned wrong content type: got %v want %v", rr.Header().Get("Content-Type"), expected)
	}

	// Проверяем, что ответ содержит JSON
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("handler returned invalid JSON: %v", err)
	}

	// Проверяем обязательные поля
	if response["status"] != "healthy" {
		t.Errorf("handler returned wrong status: got %v want %v", response["status"], "healthy")
	}

	if response["webhook"] != "https://test.com/webhook" {
		t.Errorf("handler returned wrong webhook URL: got %v want %v", response["webhook"], "https://test.com/webhook")
	}
}

func TestWebhookEndpoint(t *testing.T) {
	// Создаем мок конфигурации
	config := &Config{}

	// Создаем webhook сервер без worker pool для тестирования
	webhookServer := &WebhookServer{
		config:     config,
		webhookURL: "https://test.com",
		// Не устанавливаем bot, чтобы избежать nil pointer dereference в тестах
	}

	// Создаем тестовое сообщение
	update := tgbotapi.Update{
		UpdateID: 123456789,
		Message: &tgbotapi.Message{
			MessageID: 1,
			From: &tgbotapi.User{
				ID:        123456789,
				FirstName: "Test",
			},
			Chat: &tgbotapi.Chat{
				ID:   123456789,
				Type: "private",
			},
			Text: "https://example.com",
		},
	}

	// Сериализуем в JSON
	jsonData, err := json.Marshal(update)
	if err != nil {
		t.Fatal(err)
	}

	// Создаем тестовый запрос
	req, err := http.NewRequest("POST", "/webhook", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Создаем ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(webhookServer.handleWebhook)

	// Выполняем запрос
	handler.ServeHTTP(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Проверяем ответ
	expected := "OK"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestWebhookInvalidMethod(t *testing.T) {
	// Создаем webhook сервер
	webhookServer := &WebhookServer{
		webhookURL: "https://test.com",
	}

	// Создаем GET запрос (недопустимый метод)
	req, err := http.NewRequest("GET", "/webhook", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Создаем ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(webhookServer.handleWebhook)

	// Выполняем запрос
	handler.ServeHTTP(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

func TestWebhookInvalidContentType(t *testing.T) {
	// Создаем webhook сервер
	webhookServer := &WebhookServer{
		webhookURL: "https://test.com",
	}

	// Создаем POST запрос с неправильным Content-Type
	req, err := http.NewRequest("POST", "/webhook", bytes.NewBuffer([]byte("test")))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "text/plain")

	// Создаем ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(webhookServer.handleWebhook)

	// Выполняем запрос
	handler.ServeHTTP(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestWebhookInvalidJSON(t *testing.T) {
	// Создаем webhook сервер
	webhookServer := &WebhookServer{
		webhookURL: "https://test.com",
	}

	// Создаем POST запрос с невалидным JSON
	req, err := http.NewRequest("POST", "/webhook", bytes.NewBuffer([]byte("invalid json")))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Создаем ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(webhookServer.handleWebhook)

	// Выполняем запрос
	handler.ServeHTTP(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}
