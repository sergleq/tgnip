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
func convertToMarkdown(content *Content, originalURL string, locale Locale) string {
	var markdown strings.Builder

	// Заголовок
	markdown.WriteString(fmt.Sprintf("# %s\n\n", escapeMarkdown(content.Title)))

	// Метаинформация
	markdown.WriteString(fmt.Sprintf("%s\n\n", locale.MetadataSection))
	markdown.WriteString(fmt.Sprintf("- %s [%s](%s)\n", locale.SourceLabel, getDomain(originalURL, locale), originalURL))

	if content.Author != "" {
		markdown.WriteString(fmt.Sprintf("- %s %s\n", locale.AuthorLabel, content.Author))
	}

	if content.Date != "" {
		markdown.WriteString(fmt.Sprintf("- %s %s\n", locale.DateLabel, content.Date))
	}

	markdown.WriteString(fmt.Sprintf("- %s %s\n", locale.ProcessedLabel, time.Now().Format("2006-01-02 15:04:05")))
	markdown.WriteString("\n---\n\n")

	// Основной текст
	markdown.WriteString(fmt.Sprintf("%s\n\n", locale.ContentSection))

	// Добавляем основной контент
	markdown.WriteString(content.Markdown)
	markdown.WriteString("\n\n")

	// Футер
	markdown.WriteString("---\n\n")
	markdown.WriteString(fmt.Sprintf("%s\n", locale.FooterText))

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
func getDomain(urlStr string, locale Locale) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return locale.UnknownSource
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

// isValidFilename проверяет корректность имени файла
func isValidFilename(filename string) bool {
	// Проверяем, что имя файла не пустое
	if filename == "" || filename == ".md" {
		return false
	}

	// Проверяем, что имя файла не слишком длинное (максимум 255 символов для большинства файловых систем)
	if len(filename) > 255 {
		return false
	}

	// Проверяем, что имя файла не начинается с точки (скрытые файлы в Unix)
	if strings.HasPrefix(filename, ".") {
		return false
	}

	// Проверяем, что имя файла не заканчивается точкой (проблема в Windows)
	// Но разрешаем .md в конце
	if strings.HasSuffix(filename, ".") && !strings.HasSuffix(filename, ".md") {
		return false
	}

	// Убираем расширение для проверки
	nameWithoutExt := strings.TrimSuffix(filename, ".md")

	// Проверяем, что имя файла не содержит точку перед расширением
	if strings.HasSuffix(nameWithoutExt, ".") {
		return false
	}

	// Проверяем, что имя файла не содержит только подчеркивания (после очистки)
	if nameWithoutExt == "" || nameWithoutExt == "_" {
		return false
	}

	// Проверяем, что после удаления подчеркиваний остается что-то
	cleanName := strings.ReplaceAll(nameWithoutExt, "_", "")
	if cleanName == "" {
		return false
	}

	return true
}

// generateFilename генерирует имя файла для markdown документа
func generateFilename(originalURL string, title string) string {
	// Если есть заголовок статьи, используем его для имени файла
	if title != "" && title != "Без заголовка" {
		// Очищаем заголовок от недопустимых символов
		cleanTitle := cleanTitleForFilename(title)

		// Ограничиваем длину заголовка
		if len(cleanTitle) > 200 {
			cleanTitle = cleanTitle[:200]
		}

		// Убираем только действительно недопустимые символы для файловых систем
		// Поддерживаем кириллицу, латиницу, цифры, точки, дефисы, подчеркивания
		invalidChars := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
		filename := invalidChars.ReplaceAllString(cleanTitle, "_")

		// Добавляем расширение
		filename = filename + ".md"

		// Проверяем корректность имени файла
		if !isValidFilename(filename) {
			return "new_markdown_file.md"
		}

		return filename
	}

	// Если заголовка нет, используем URL как fallback
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
	if len(urlPath) > 150 {
		urlPath = urlPath[:150]
	}

	// Добавляем домен и расширение
	domain := u.Hostname()
	domain = strings.ReplaceAll(domain, ".", "_")

	filename := fmt.Sprintf("%s_%s", domain, urlPath)

	// Убираем только действительно недопустимые символы для файловых систем
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	filename = invalidChars.ReplaceAllString(filename, "_")

	// Добавляем расширение
	filename = filename + ".md"

	// Проверяем корректность имени файла
	if !isValidFilename(filename) {
		return "new_markdown_file.md"
	}

	return filename
}

// cleanTitleForFilename очищает заголовок для использования в имени файла
func cleanTitleForFilename(title string) string {
	// Убираем лишние пробелы
	title = strings.TrimSpace(title)

	// Заменяем пробелы на подчеркивания
	title = strings.ReplaceAll(title, " ", "_")

	// Убираем только действительно недопустимые символы для файловых систем
	// Поддерживаем кириллицу, латиницу, цифры, точки, запятые, восклицательные знаки и другие
	// Заменяем только символы, которые запрещены в именах файлов
	title = strings.ReplaceAll(title, "<", "_")
	title = strings.ReplaceAll(title, ">", "_")
	title = strings.ReplaceAll(title, ":", "_")
	title = strings.ReplaceAll(title, "\"", "_")
	title = strings.ReplaceAll(title, "'", "_")
	title = strings.ReplaceAll(title, "/", "_")
	title = strings.ReplaceAll(title, "\\", "_")
	title = strings.ReplaceAll(title, "|", "_")
	title = strings.ReplaceAll(title, "?", "_")
	title = strings.ReplaceAll(title, "*", "_")

	// Убираем множественные подчеркивания
	title = regexp.MustCompile(`_+`).ReplaceAllString(title, "_")

	// Убираем подчеркивания в начале и конце
	title = strings.Trim(title, "_")

	return title
}
