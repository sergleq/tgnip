package internal

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestGetLocale(t *testing.T) {
	tests := []struct {
		name           string
		languageCode   string
		expectedLocale string
	}{
		{
			name:           "Russian language",
			languageCode:   "ru",
			expectedLocale: "ru",
		},
		{
			name:           "English language",
			languageCode:   "en",
			expectedLocale: "en",
		},
		{
			name:           "English US variant",
			languageCode:   "en-US",
			expectedLocale: "en",
		},
		{
			name:           "Unknown language defaults to English",
			languageCode:   "unknown",
			expectedLocale: "en",
		},
		{
			name:           "Empty language defaults to English",
			languageCode:   "",
			expectedLocale: "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := &tgbotapi.Message{
				From: &tgbotapi.User{
					LanguageCode: tt.languageCode,
				},
			}

			locale := GetLocale(message)
			expectedLocale := locales[tt.expectedLocale]

			// Проверяем, что полученная локализация соответствует ожидаемой
			if locale.WelcomeMessage != expectedLocale.WelcomeMessage {
				t.Errorf("Expected welcome message for %s, got different locale", tt.expectedLocale)
			}

			if locale.ProcessingMessage != expectedLocale.ProcessingMessage {
				t.Errorf("Expected processing message for %s, got different locale", tt.expectedLocale)
			}
		})
	}
}

func TestGetSupportedLanguages(t *testing.T) {
	languages := getSupportedLanguages()

	// Проверяем, что все поддерживаемые языки присутствуют
	expectedLanguages := []string{"ru", "en"}

	if len(languages) != len(expectedLanguages) {
		t.Errorf("Expected %d languages, got %d", len(expectedLanguages), len(languages))
	}

	// Проверяем, что все ожидаемые языки присутствуют
	for _, expectedLang := range expectedLanguages {
		found := false
		for _, lang := range languages {
			if lang == expectedLang {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected language %s not found in supported languages", expectedLang)
		}
	}
}

func TestLocaleStructure(t *testing.T) {
	// Проверяем, что все локали имеют все необходимые поля
	for lang, locale := range locales {
		if locale.WelcomeMessage == "" {
			t.Errorf("WelcomeMessage is empty for language %s", lang)
		}
		if locale.ProcessingMessage == "" {
			t.Errorf("ProcessingMessage is empty for language %s", lang)
		}
		if locale.InvalidURLMessage == "" {
			t.Errorf("InvalidURLMessage is empty for language %s", lang)
		}
		if locale.ErrorProcessingMsg == "" {
			t.Errorf("ErrorProcessingMsg is empty for language %s", lang)
		}
		if locale.ErrorSendingMsg == "" {
			t.Errorf("ErrorSendingMsg is empty for language %s", lang)
		}
		if locale.SuccessMessage == "" {
			t.Errorf("SuccessMessage is empty for language %s", lang)
		}
	}
}

func TestDefaultLocale(t *testing.T) {
	// Проверяем, что при отсутствии языка пользователя возвращается английский
	message := &tgbotapi.Message{
		From: &tgbotapi.User{
			LanguageCode: "",
		},
	}

	locale := GetLocale(message)
	expectedLocale := locales["en"]

	if locale.WelcomeMessage != expectedLocale.WelcomeMessage {
		t.Error("Default locale should be English")
	}
}

func TestNilUser(t *testing.T) {
	// Проверяем, что при отсутствии информации о пользователе возвращается английский
	message := &tgbotapi.Message{
		From: nil,
	}

	locale := GetLocale(message)
	expectedLocale := locales["en"]

	if locale.WelcomeMessage != expectedLocale.WelcomeMessage {
		t.Error("Should return English locale when user is nil")
	}
}
