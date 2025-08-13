package main

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Locale представляет локализацию для определенного языка
type Locale struct {
	WelcomeMessage     string
	ProcessingMessage  string
	InvalidURLMessage  string
	ErrorProcessingMsg string
	ErrorSendingMsg    string
	SuccessMessage     string
}

// locales содержит все поддерживаемые языки
var locales = map[string]Locale{
	"ru": {
		WelcomeMessage: `Привет! Я бот для преобразования ссылок в markdown файлы.

Отправьте мне ссылку на веб-страницу, и я создам для вас markdown файл с очищенным содержимым.

Поддерживаемые форматы:
- Статьи и новости
- Блог-посты
- Документация

Просто отправьте ссылку!`,
		ProcessingMessage:  "⏳ Обрабатываю ссылку...",
		InvalidURLMessage:  "Пожалуйста, отправьте валидную ссылку на веб-страницу.",
		ErrorProcessingMsg: "❌ Не удалось обработать ссылку. Проверьте, что ссылка корректна и доступна.",
		ErrorSendingMsg:    "❌ Не удалось отправить файл.",
		SuccessMessage:     "✅ Файл успешно создан!",
	},
	"en": {
		WelcomeMessage: `Hello! I'm a bot for converting links to markdown files.

Send me a link to a web page, and I'll create a markdown file with cleaned content for you.

Supported formats:
- Articles and news
- Blog posts
- Documentation

Just send a link!`,
		ProcessingMessage:  "⏳ Processing link...",
		InvalidURLMessage:  "Please send a valid link to a web page.",
		ErrorProcessingMsg: "❌ Failed to process the link. Check that the link is correct and accessible.",
		ErrorSendingMsg:    "❌ Failed to send the file.",
		SuccessMessage:     "✅ File successfully created!",
	},
}

// getLocale определяет язык пользователя и возвращает соответствующую локализацию
func getLocale(message *tgbotapi.Message) Locale {
	// Проверяем язык пользователя
	if message.From != nil && message.From.LanguageCode != "" {
		lang := strings.ToLower(message.From.LanguageCode)

		// Проверяем точное совпадение
		if locale, exists := locales[lang]; exists {
			return locale
		}

		// Проверяем основную часть языка (например, "en" для "en-US")
		if len(lang) >= 2 {
			mainLang := lang[:2]
			if locale, exists := locales[mainLang]; exists {
				return locale
			}
		}
	}

	// Возвращаем английский как язык по умолчанию
	return locales["en"]
}

// getSupportedLanguages возвращает список поддерживаемых языков
func getSupportedLanguages() []string {
	languages := make([]string, 0, len(locales))
	for lang := range locales {
		languages = append(languages, lang)
	}
	return languages
}
