package main

import (
	"os"
	"strings"

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

	// Загрузка переменных окружения
	if err := godotenv.Load(); err != nil {
		logger.Warn("Файл .env не найден, используем переменные окружения системы")
	}
}

func main() {
	// Загрузка конфигурации
	config := internal.LoadConfig()

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

	// Создание бота
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		logger.Fatal(err)
	}

	bot.Debug = false
	logger.Infof("Бот %s запущен", bot.Self.UserName)

	// Проверяем режим работы
	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL != "" {
		// Webhook режим
		logger.Info("Запуск в webhook режиме")
		webhookServer := internal.NewWebhookServer(bot, config)

		if err := webhookServer.Start(); err != nil {
			logger.Fatal("Ошибка запуска webhook сервера: ", err)
		}
		defer webhookServer.Stop()

		// Если это HTTPS URL, устанавливаем webhook автоматически
		if strings.HasPrefix(webhookURL, "https://") {
			if err := webhookServer.SetWebhook(); err != nil {
				logger.Fatal("Ошибка установки webhook: ", err)
			}
		}

		// Ждем сигнала завершения
		select {}
	} else {
		// Polling режим
		logger.Info("Запуск в polling режиме")

		// Настройка обновлений
		updateConfig := tgbotapi.NewUpdate(0)
		updateConfig.Timeout = 60

		updates := bot.GetUpdatesChan(updateConfig)

		// Обработка сообщений
		for update := range updates {
			if update.Message == nil {
				continue
			}

			// Обрабатываем сообщение напрямую
			internal.HandleMessage(bot, update.Message)
		}
	}
}
