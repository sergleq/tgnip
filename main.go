package main

import (
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func init() {
	// Инициализация логгера
	logger = logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Загрузка переменных окружения
	if err := godotenv.Load(); err != nil {
		logger.Warn("Файл .env не найден, используем переменные окружения системы")
	}
}

func main() {
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

	// Настройка обновлений
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	// Обработка сообщений
	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Обработка команды /start
		if update.Message.IsCommand() && update.Message.Command() == "start" {
			handleStartCommand(bot, update.Message)
			continue
		}

		// Обработка ссылок
		if update.Message.Text != "" {
			handleURLMessage(bot, update.Message)
		}
	}
}

func handleStartCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	locale := getLocale(message)
	msg := tgbotapi.NewMessage(message.Chat.ID, locale.WelcomeMessage)
	bot.Send(msg)
}

func handleURLMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	locale := getLocale(message)
	url := message.Text

	// Проверка, что это действительно ссылка
	if !isValidURL(url) {
		msg := tgbotapi.NewMessage(message.Chat.ID, locale.InvalidURLMessage)
		bot.Send(msg)
		return
	}

	// Отправляем сообщение о начале обработки
	processingMsg := tgbotapi.NewMessage(message.Chat.ID, locale.ProcessingMessage)
	sentMsg, err := bot.Send(processingMsg)
	if err != nil {
		logger.Errorf("Ошибка при отправке сообщения о обработке: %v", err)
	}

	// Извлекаем контент
	content, err := extractContent(url)
	if err != nil {
		logger.Errorf("Ошибка при извлечении контента: %v", err)
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, locale.ErrorProcessingMsg)
		bot.Send(errorMsg)
		return
	}

	// Конвертируем в markdown
	markdown := convertToMarkdown(content, url)

	// Создаем файл
	filename := generateFilename(url)
	file := tgbotapi.NewDocument(message.Chat.ID, tgbotapi.FileBytes{
		Name:  filename,
		Bytes: []byte(markdown),
	})

	// Отправляем файл
	if _, err := bot.Send(file); err != nil {
		logger.Errorf("Ошибка при отправке файла: %v", err)
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, locale.ErrorSendingMsg)
		bot.Send(errorMsg)
		return
	}

	// Удаляем сообщение о обработке
	if sentMsg.MessageID != 0 {
		deleteMsg := tgbotapi.NewDeleteMessage(message.Chat.ID, sentMsg.MessageID)
		bot.Send(deleteMsg)
	}

	logger.Infof("Файл успешно отправлен пользователю %d", message.Chat.ID)
}
