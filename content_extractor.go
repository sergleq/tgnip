package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	readability "github.com/go-shiori/go-readability"
)

// Content представляет извлеченный контент
type Content struct {
	Title    string
	Markdown string
	URL      string
	Author   string
	Date     string
}

// extractContent извлекает контент из веб-страницы
func extractContent(pageURL string) (*Content, error) {
	fmt.Printf("Извлекаю контент из: %s\n", pageURL)

	// Создаем HTTP клиент с таймаутом
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Выполняем запрос
	resp, err := client.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка при загрузке страницы: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("статус ответа: %d", resp.StatusCode)
	}

	// Читаем тело ответа
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении тела ответа: %w", err)
	}

	// Парсим HTML с помощью goquery для извлечения метаданных
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("ошибка при парсинге HTML: %w", err)
	}

	// Парсим URL для go-readability
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка при парсинге URL: %w", err)
	}

	// Извлекаем контент с помощью go-readability
	article, err := readability.FromReader(strings.NewReader(string(bodyBytes)), parsedURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка при извлечении контента: %w", err)
	}

	// Создаем объект контента
	content := &Content{
		URL: pageURL,
	}

	// Извлекаем заголовок (приоритет go-readability, затем fallback)
	if article.Title != "" {
		content.Title = article.Title
	} else {
		content.Title = extractTitle(doc)
	}

	// Извлекаем основной текст и конвертируем в markdown
	content.Markdown = extractAndConvertToMarkdown(article)

	// Извлекаем автора
	if article.Byline != "" {
		content.Author = article.Byline
	} else {
		content.Author = extractAuthor(doc)
	}

	// Извлекаем дату
	content.Date = extractDate(doc)

	fmt.Printf("Извлечено: заголовок='%s', длина markdown=%d символов\n",
		content.Title, len(content.Markdown))

	return content, nil
}

// extractTitle извлекает заголовок страницы
func extractTitle(doc *goquery.Document) string {
	// Пробуем различные селекторы для заголовка
	selectors := []string{
		"h1",
		"title",
		"[property='og:title']",
		"[name='twitter:title']",
		".post-title",
		".article-title",
		".entry-title",
	}

	for _, selector := range selectors {
		if selector == "title" {
			title := doc.Find("title").Text()
			if title != "" {
				return strings.TrimSpace(title)
			}
		} else if strings.Contains(selector, "property=") || strings.Contains(selector, "name=") {
			title := doc.Find(selector).AttrOr("content", "")
			if title != "" {
				return strings.TrimSpace(title)
			}
		} else {
			title := doc.Find(selector).First().Text()
			if title != "" {
				return strings.TrimSpace(title)
			}
		}
	}

	return "Без заголовка"
}

// extractMainText извлекает основной текст статьи
func extractMainText(doc *goquery.Document) string {
	// Удаляем ненужные элементы
	doc.Find("script, style, nav, header, footer, .sidebar, .advertisement, .ads, .comments, .social-share").Remove()

	// Пробуем найти основной контент
	selectors := []string{
		"article",
		".post-content",
		".article-content",
		".entry-content",
		".content",
		"main",
		".main-content",
		"#content",
		".post-body",
		".article-body",
	}

	for _, selector := range selectors {
		content := doc.Find(selector).First()
		if content.Length() > 0 {
			text := content.Text()
			if len(text) > 100 { // Минимальная длина для валидного контента
				return text
			}
		}
	}

	// Если не нашли основной контент, берем весь body
	body := doc.Find("body").Text()
	return body
}

// extractAuthor извлекает автора статьи
func extractAuthor(doc *goquery.Document) string {
	selectors := []string{
		".author",
		".post-author",
		".article-author",
		"[rel='author']",
		"[property='article:author']",
		".byline",
		".author-name",
	}

	for _, selector := range selectors {
		author := doc.Find(selector).First().Text()
		if author != "" {
			return strings.TrimSpace(author)
		}
	}

	return ""
}

// extractDate извлекает дату публикации
func extractDate(doc *goquery.Document) string {
	selectors := []string{
		".date",
		".post-date",
		".article-date",
		".published",
		"[property='article:published_time']",
		"[property='og:published_time']",
		"time",
		".timestamp",
	}

	for _, selector := range selectors {
		date := doc.Find(selector).First()
		if selector == "time" {
			dateStr := date.AttrOr("datetime", "")
			if dateStr == "" {
				dateStr = date.Text()
			}
			if dateStr != "" {
				return strings.TrimSpace(dateStr)
			}
		} else if strings.Contains(selector, "property=") {
			dateStr := date.AttrOr("content", "")
			if dateStr != "" {
				return strings.TrimSpace(dateStr)
			}
		} else {
			dateStr := date.Text()
			if dateStr != "" {
				return strings.TrimSpace(dateStr)
			}
		}
	}

	return ""
}

// cleanText очищает текст от лишних символов
func cleanText(text string) string {
	// Удаляем пробелы в начале и конце
	text = strings.TrimSpace(text)

	// Разбиваем на абзацы и очищаем каждый
	paragraphs := strings.Split(text, "\n")
	var cleanParagraphs []string

	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		// Удаляем множественные пробелы внутри абзаца
		re := regexp.MustCompile(`\s+`)
		p = re.ReplaceAllString(p, " ")
		if len(p) > 10 { // Минимальная длина абзаца
			cleanParagraphs = append(cleanParagraphs, p)
		}
	}

	return strings.Join(cleanParagraphs, "\n\n")
}

// isValidURL проверяет, является ли строка валидной URL
func isValidURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// extractAndConvertToMarkdown извлекает контент и конвертирует его в markdown
func extractAndConvertToMarkdown(article readability.Article) string {
	if article.Content == "" {
		return ""
	}

	// Создаем конвертер html-to-markdown с настройками
	converter := md.NewConverter("", true, nil)

	// Добавляем правило для лучшей обработки таблиц
	converter.AddRules(
		md.Rule{
			Filter: []string{"table"},
			Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
				return convertTableToMarkdown(selec)
			},
		},
	)

	// Конвертируем HTML в markdown
	markdown, err := converter.ConvertString(article.Content)
	if err != nil {
		// Fallback к простому тексту
		return cleanText(article.TextContent)
	}

	// Обрабатываем блоки кода после конвертации
	markdown = processCodeBlocksInMarkdown(markdown, article.Content)

	// Очищаем результат
	markdown = strings.TrimSpace(markdown)

	// Убираем лишние пустые строки, но сохраняем структуру
	lines := strings.Split(markdown, "\n")
	var cleanLines []string
	var lastWasEmpty bool

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			// Добавляем пустую строку только если предыдущая не была пустой
			if !lastWasEmpty {
				cleanLines = append(cleanLines, "")
				lastWasEmpty = true
			}
		} else {
			cleanLines = append(cleanLines, line)
			lastWasEmpty = false
		}
	}

	// Убираем пустые строки в начале и конце
	result := strings.Join(cleanLines, "\n")
	result = strings.Trim(result, "\n")

	return result
}

// processCodeBlocksInMarkdown обрабатывает блоки кода в markdown, исправляя языки программирования
func processCodeBlocksInMarkdown(markdown, originalHTML string) string {
	// Парсим оригинальный HTML для извлечения информации о языках
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(originalHTML))
	if err != nil {
		return markdown
	}

	// Находим все pre блоки и их языки
	var codeBlocks []struct {
		content string
		lang    string
	}

	doc.Find("pre").Each(func(i int, s *goquery.Selection) {
		code := s.Find("code").Text()
		if code == "" {
			code = s.Text()
		}

		// Определяем язык программирования
		lang := extractLanguageFromCode(s.Find("code"))

		codeBlocks = append(codeBlocks, struct {
			content string
			lang    string
		}{content: code, lang: lang})
	})

	// Исправляем блоки кода с неправильными языками
	result := markdown

	// Сначала исправляем блоки с неправильными языками (например, "highlight javascript")
	// Используем регулярное выражение для поиска блоков кода
	re := regexp.MustCompile("```([a-zA-Z0-9_-]+)\\s+([a-zA-Z0-9_-]+)\\n")
	matches := re.FindAllStringSubmatch(result, -1)

	for _, match := range matches {
		if len(match) == 3 {
			// match[1] - первый язык (например, "highlight")
			// match[2] - второй язык (например, "javascript")
			firstLang := match[1]
			secondLang := match[2]

			// Проверяем, является ли второй язык правильным
			normalizedLang := normalizeLanguageName(secondLang)
			if normalizedLang != "" {
				// Заменяем "highlight javascript" на "javascript"
				oldBlock := fmt.Sprintf("```%s %s\n", firstLang, secondLang)
				newBlock := fmt.Sprintf("```%s\n", normalizedLang)
				result = strings.Replace(result, oldBlock, newBlock, 1)
			}
		}
	}

	// Затем обрабатываем блоки без языка (если они есть)
	for _, block := range codeBlocks {
		if block.lang != "" {
			// Ищем блок кода с этим содержимым без языка
			standardBlock := fmt.Sprintf("```\n%s\n```", block.content)
			langBlock := fmt.Sprintf("```%s\n%s\n```", block.lang, block.content)
			result = strings.Replace(result, standardBlock, langBlock, 1)
		}
	}

	return result
}

// extractLanguageFromCode извлекает язык программирования из тега code
func extractLanguageFromCode(codeElement *goquery.Selection) string {
	if codeElement.Length() == 0 {
		return ""
	}

	// Получаем все классы
	classAttr := codeElement.AttrOr("class", "")
	if classAttr == "" {
		return ""
	}

	// Разбиваем классы на отдельные части
	classes := strings.Fields(classAttr)

	// Сначала ищем префиксы language- и lang-
	for _, class := range classes {
		class = strings.TrimSpace(class)

		if strings.HasPrefix(class, "language-") {
			// Формат: language-python, language-javascript
			lang := strings.TrimPrefix(class, "language-")
			normalized := normalizeLanguageName(lang)
			if normalized != "" {
				return normalized
			}
		}

		if strings.HasPrefix(class, "lang-") {
			// Формат: lang-python, lang-javascript
			lang := strings.TrimPrefix(class, "lang-")
			normalized := normalizeLanguageName(lang)
			if normalized != "" {
				return normalized
			}
		}
	}

	// Если не нашли префиксы, проверяем прямые названия языков
	// Но исключаем общие CSS классы, которые не являются языками
	excludedClasses := map[string]bool{
		"highlight":   true,
		"code":        true,
		"source":      true,
		"source-code": true,
		"prettyprint": true,
		"linenums":    true,
		"hljs":        true,
		"language":    true,
		"lang":        true,
	}

	for _, class := range classes {
		class = strings.TrimSpace(class)

		// Пропускаем исключенные классы
		if excludedClasses[class] {
			continue
		}

		lang := normalizeLanguageName(class)
		if lang != "" {
			return lang
		}
	}

	return ""
}

// normalizeLanguageName нормализует названия языков программирования
func normalizeLanguageName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))

	// Маппинг различных вариантов названий языков
	languageMap := map[string]string{
		// JavaScript
		"js":         "javascript",
		"javascript": "javascript",
		"ecmascript": "javascript",

		// Python
		"python": "python",
		"py":     "python",

		// Java
		"java": "java",

		// C/C++
		"c":         "c",
		"cpp":       "cpp",
		"c++":       "cpp",
		"cplusplus": "cpp",

		// C#
		"c#":     "csharp",
		"csharp": "csharp",
		"cs":     "csharp",

		// Go
		"go":     "go",
		"golang": "go",

		// Rust
		"rust": "rust",
		"rs":   "rust",

		// PHP
		"php": "php",

		// Ruby
		"ruby": "ruby",
		"rb":   "ruby",

		// Swift
		"swift": "swift",

		// Kotlin
		"kotlin": "kotlin",
		"kt":     "kotlin",

		// TypeScript
		"typescript": "typescript",
		"ts":         "typescript",

		// HTML
		"html": "html",
		"htm":  "html",

		// CSS
		"css": "css",

		// SQL
		"sql": "sql",

		// Shell
		"bash":  "bash",
		"shell": "bash",
		"sh":    "bash",
		"zsh":   "bash",

		// JSON
		"json": "json",

		// XML
		"xml": "xml",

		// YAML
		"yaml": "yaml",
		"yml":  "yaml",

		// Markdown
		"markdown": "markdown",
		"md":       "markdown",

		// Docker
		"dockerfile": "dockerfile",
		"docker":     "dockerfile",

		// Git
		"git": "git",

		// Diff
		"diff":  "diff",
		"patch": "diff",
	}

	if normalized, exists := languageMap[name]; exists {
		return normalized
	}

	// Если не нашли в маппинге, возвращаем как есть (если это похоже на язык)
	if len(name) > 0 && len(name) <= 20 && !strings.Contains(name, " ") {
		return name
	}

	return ""
}

// convertTableToMarkdown конвертирует HTML таблицу в markdown формат
func convertTableToMarkdown(table *goquery.Selection) *string {
	var rows [][]string
	var headers []string

	// Обрабатываем заголовки (thead)
	table.Find("thead tr").Each(func(i int, tr *goquery.Selection) {
		var row []string
		tr.Find("th, td").Each(func(j int, cell *goquery.Selection) {
			text := strings.TrimSpace(cell.Text())
			row = append(row, text)
		})
		if len(row) > 0 {
			headers = row
		}
	})

	// Если нет thead, берем первую строку как заголовки
	if len(headers) == 0 {
		table.Find("tr").First().Each(func(i int, tr *goquery.Selection) {
			tr.Find("th, td").Each(func(j int, cell *goquery.Selection) {
				text := strings.TrimSpace(cell.Text())
				headers = append(headers, text)
			})
		})
	}

	// Обрабатываем строки данных (tbody или все tr кроме первой)
	table.Find("tbody tr, tr").Each(func(i int, tr *goquery.Selection) {
		// Пропускаем строку заголовков, если она уже обработана
		if len(headers) > 0 && i == 0 && tr.Find("th").Length() > 0 {
			return
		}

		var row []string
		tr.Find("td, th").Each(func(j int, cell *goquery.Selection) {
			text := strings.TrimSpace(cell.Text())
			row = append(row, text)
		})
		if len(row) > 0 {
			rows = append(rows, row)
		}
	})

	// Если нет данных, возвращаем nil
	if len(headers) == 0 && len(rows) == 0 {
		return nil
	}

	// Строим markdown таблицу
	var result strings.Builder

	// Заголовки
	if len(headers) > 0 {
		result.WriteString("| ")
		for _, header := range headers {
			result.WriteString(escapeMarkdownTableCell(header))
			result.WriteString(" | ")
		}
		result.WriteString("\n")

		// Разделитель
		result.WriteString("| ")
		for range headers {
			result.WriteString("--- | ")
		}
		result.WriteString("\n")
	}

	// Строки данных
	for i, row := range rows {
		result.WriteString("| ")
		for _, cell := range row {
			result.WriteString(escapeMarkdownTableCell(cell))
			result.WriteString(" | ")
		}
		result.WriteString("\n")

		// Добавляем пустую строку после таблицы только если это последняя строка
		if i == len(rows)-1 {
			result.WriteString("\n")
		}
	}

	markdown := result.String()
	return &markdown
}

// escapeMarkdownTableCell экранирует специальные символы в ячейках таблицы
func escapeMarkdownTableCell(text string) string {
	// Заменяем символы, которые могут сломать таблицу
	text = strings.ReplaceAll(text, "|", "\\|")
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")

	// Убираем лишние пробелы
	text = strings.TrimSpace(text)

	return text
}
