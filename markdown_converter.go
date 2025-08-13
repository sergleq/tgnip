package main

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"
)

// convertToMarkdown конвертирует контент в markdown формат
func convertToMarkdown(content *Content, originalURL string) string {
	var markdown strings.Builder

	// Заголовок
	markdown.WriteString(fmt.Sprintf("# %s\n\n", escapeMarkdown(content.Title)))

	// Метаинформация
	markdown.WriteString("## Метаинформация\n\n")
	markdown.WriteString(fmt.Sprintf("- **Источник:** [%s](%s)\n", getDomain(originalURL), originalURL))

	if content.Author != "" {
		markdown.WriteString(fmt.Sprintf("- **Автор:** %s\n", content.Author))
	}

	if content.Date != "" {
		markdown.WriteString(fmt.Sprintf("- **Дата:** %s\n", content.Date))
	}

	markdown.WriteString(fmt.Sprintf("- **Обработано:** %s\n", time.Now().Format("2006-01-02 15:04:05")))
	markdown.WriteString("\n---\n\n")

	// Основной текст
	markdown.WriteString("## Содержание\n\n")

	// Добавляем основной контент
	markdown.WriteString(content.Markdown)
	markdown.WriteString("\n\n")

	// Футер
	markdown.WriteString("---\n\n")
	markdown.WriteString("*Этот документ был автоматически создан ботом для преобразования веб-страниц в markdown формат.*\n")

	return markdown.String()
}

// escapeMarkdown экранирует специальные символы markdown
func escapeMarkdown(text string) string {
	// Экранируем символы, которые имеют специальное значение в markdown
	escaped := text
	escaped = strings.ReplaceAll(escaped, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "*", "\\*")
	escaped = strings.ReplaceAll(escaped, "_", "\\_")
	escaped = strings.ReplaceAll(escaped, "`", "\\`")
	escaped = strings.ReplaceAll(escaped, "#", "\\#")
	escaped = strings.ReplaceAll(escaped, "+", "\\+")
	escaped = strings.ReplaceAll(escaped, "-", "\\-")
	escaped = strings.ReplaceAll(escaped, ".", "\\.")
	escaped = strings.ReplaceAll(escaped, "!", "\\!")
	escaped = strings.ReplaceAll(escaped, "[", "\\[")
	escaped = strings.ReplaceAll(escaped, "]", "\\]")
	escaped = strings.ReplaceAll(escaped, "(", "\\(")
	escaped = strings.ReplaceAll(escaped, ")", "\\)")

	return escaped
}

// getDomain извлекает домен из URL
func getDomain(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "Неизвестный источник"
	}
	return u.Hostname()
}

// isHeading проверяет, является ли текст заголовком
func isHeading(text string) bool {
	// Проверяем, начинается ли текст с цифры и точки (например, "1. Заголовок")
	headingPattern := regexp.MustCompile(`^\d+\.\s+`)
	return headingPattern.MatchString(text) ||
		len(text) < 100 && // Короткий текст
			!strings.Contains(text, ".") && // Без точек
			!strings.Contains(text, ",") && // Без запятых
			len(strings.Split(text, " ")) < 10 // Меньше 10 слов
}

// generateFilename генерирует имя файла для markdown документа
func generateFilename(originalURL string) string {
	u, err := url.Parse(originalURL)
	if err != nil {
		return fmt.Sprintf("article_%d.md", time.Now().Unix())
	}

	// Извлекаем путь из URL
	urlPath := u.Path
	if urlPath == "" || urlPath == "/" {
		urlPath = "index"
	}

	// Убираем расширение файла
	urlPath = strings.TrimSuffix(urlPath, path.Ext(urlPath))

	// Заменяем слеши на подчеркивания
	urlPath = strings.ReplaceAll(urlPath, "/", "_")
	urlPath = strings.ReplaceAll(urlPath, "-", "_")

	// Убираем лишние подчеркивания
	urlPath = strings.Trim(urlPath, "_")

	// Ограничиваем длину
	if len(urlPath) > 50 {
		urlPath = urlPath[:50]
	}

	// Добавляем домен и расширение
	domain := u.Hostname()
	domain = strings.ReplaceAll(domain, ".", "_")

	filename := fmt.Sprintf("%s_%s.md", domain, urlPath)

	// Убираем недопустимые символы
	invalidChars := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	filename = invalidChars.ReplaceAllString(filename, "_")

	return filename
}
