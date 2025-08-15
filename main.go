package main

import (
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"tgnip/internal"
)

var logger *logrus.Logger

func init() {
	// Инициализация логгера
	logger = logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Устанавливаем логгер для internal пакета
	internal.SetLogger(logger)

	// Загрузка переменных окружения (только если они не установлены)
	if err := godotenv.Load(); err != nil {
		logger.Warn("Файл .env не найден, используем переменные окружения системы")
	} else {
		logger.Info("Переменные окружения загружены из .env файла")
	}
}

func main() {
	// Загрузка конфигурации
	config := internal.LoadConfig()

	var bot *tgbotapi.BotAPI
	var err error

	// Настройка уровня логирования
	switch config.LogLevel {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	// Получение токена бота
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		logger.Fatal("TELEGRAM_BOT_TOKEN не установлен")
	}

	// Получение URL для webhook
	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		logger.Fatal("WEBHOOK_URL не установлен. Бот работает только в webhook режиме")
	}

	// Проверяем, не является ли токен тестовым
	if token == "test_token" {
		logger.Info("Тестовый токен, пропускаем создание бота")
		bot = nil
	} else {
		// Создание бота
		bot, err = tgbotapi.NewBotAPI(token)
		if err != nil {
			logger.Fatal(err)
		}

		bot.Debug = false
		logger.Infof("Бот %s запущен", bot.Self.UserName)
	}

	// Запуск webhook сервера
	logger.Info("Запуск webhook сервера")
	webhookServer := internal.NewWebhookServer(bot, config)

	if err := webhookServer.Start(); err != nil {
		logger.Fatal("Ошибка запуска webhook сервера: ", err)
	}
	defer webhookServer.Stop()

	// Webhook должен быть установлен вручную через curl или Telegram API
	// Код не устанавливает webhook автоматически
	logger.Info("Webhook не устанавливается автоматически. Установите webhook вручную:")
	logger.Infof("curl -X POST https://api.telegram.org/bot<TOKEN>/setWebhook \\")
	logger.Infof("  -H 'Content-Type: application/json' \\")
	logger.Infof("  -d '{\"url\": \"%s/telegram/webhook\", \"secret_token\": \"%s\"}'", webhookURL, os.Getenv("WEBHOOK_SECRET_TOKEN"))

	// Ждем сигнала завершения
	select {}
}
