package internal

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

// SetLogger устанавливает логгер для пакета
func SetLogger(l *logrus.Logger) {
	logger = l
}

// HandleMessage обрабатывает входящие сообщения
func HandleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	// Обработка команды /start
	if message.IsCommand() && message.Command() == "start" {
		HandleStartCommand(bot, message)
		return
	}

	// Обработка ссылок
	if message.Text != "" {
		HandleURLMessage(bot, message)
	}
}

// HandleStartCommand обрабатывает команду /start
func HandleStartCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	locale := GetLocale(message)
	msg := tgbotapi.NewMessage(message.Chat.ID, locale.WelcomeMessage)
	bot.Send(msg)
}

// HandleURLMessage обрабатывает сообщения с URL
func HandleURLMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	locale := GetLocale(message)
	url := message.Text

	// Проверка, что это действительно ссылка
	if !IsValidURL(url) {
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
	content, err := ExtractContent(url)
	if err != nil {
		logger.Errorf("Ошибка при извлечении контента: %v", err)
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, locale.ErrorProcessingMsg)
		bot.Send(errorMsg)
		return
	}

	// Конвертируем в markdown
	markdown := ConvertToMarkdown(content, url, locale)

	// Создаем файл
	filename := GenerateFilename(url, content.Title)
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
