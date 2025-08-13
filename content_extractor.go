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

	// ИЗВЛЕКАЕМ ЯЗЫКИ ПРОГРАММИРОВАНИЯ ИЗ СЫРОГО HTML ДО ОБРАБОТКИ go-readability
	languages := extractCodeLanguagesInOrder(string(bodyBytes))

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

	// Извлекаем основной текст и конвертируем в markdown с сохраненными языками
	content.Markdown = extractAndConvertToMarkdown(article, languages)

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
func extractAndConvertToMarkdown(article readability.Article, languages []string) string {
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

	// Применяем языки программирования к блокам кода в markdown по порядку
	markdown = applyLanguagesToMarkdownInOrder(markdown, languages)

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

// extractCodeLanguagesInOrder извлекает языки программирования из HTML в порядке их появления
func extractCodeLanguagesInOrder(htmlContent string) []string {
	var languages []string

	// Парсим HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return languages
	}

	// Обрабатываем все pre блоки
	doc.Find("pre").Each(func(i int, s *goquery.Selection) {
		codeElement := s.Find("code")
		if codeElement.Length() == 0 {
			return
		}

		// Извлекаем язык программирования
		language := extractLanguageFromCode(codeElement)
		languages = append(languages, language)
	})

	// Обрабатываем также одиночные теги code (не в pre)
	doc.Find("code").Each(func(i int, s *goquery.Selection) {
		// Проверяем, что это не внутри pre (чтобы не обрабатывать дважды)
		if s.Parent().Is("pre") {
			return
		}

		// Извлекаем язык программирования
		language := extractLanguageFromCode(s)
		languages = append(languages, language)
	})

	return languages
}

// applyLanguagesToMarkdownInOrder применяет языки программирования к блокам кода в markdown по порядку
func applyLanguagesToMarkdownInOrder(markdown string, languages []string) string {
	// Регулярное выражение для поиска блоков кода (с языком или без)
	// Ищем ``` в начале строки, затем содержимое, затем ``` в конце
	re := regexp.MustCompile(`(?m)^` + "```" + `[^\n]*\n((?:.*\n)*?.*?)\n` + "```" + `\s*$`)

	languageIndex := 0

	return re.ReplaceAllStringFunc(markdown, func(match string) string {
		// Извлекаем содержимое кода из блока
		codeMatch := re.FindStringSubmatch(match)
		if len(codeMatch) < 2 {
			return match
		}

		codeContent := codeMatch[1]

		// Применяем язык по порядку, если он есть
		if languageIndex < len(languages) && languages[languageIndex] != "" {
			language := languages[languageIndex]
			languageIndex++
			return fmt.Sprintf("```%s\n%s\n```", language, codeContent)
		}

		languageIndex++
		return match
	})
}

// processCodeBlocksInHTML обрабатывает блоки кода в HTML до конвертации в markdown
func processCodeBlocksInHTML(htmlContent string) string {
	// Парсим HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent
	}

	// Обрабатываем все pre блоки
	doc.Find("pre").Each(func(i int, s *goquery.Selection) {
		codeElement := s.Find("code")
		if codeElement.Length() == 0 {
			return
		}

		// Извлекаем язык программирования
		language := extractLanguageFromCode(codeElement)

		if language != "" {
			// Добавляем атрибут data-language к тегу pre для лучшей обработки
			s.SetAttr("data-language", language)

			// Также добавляем класс language-{lang} к тегу code, если его нет
			currentClass := codeElement.AttrOr("class", "")
			if !strings.Contains(currentClass, "language-"+language) {
				newClass := currentClass
				if newClass != "" {
					newClass += " "
				}
				newClass += "language-" + language
				codeElement.SetAttr("class", newClass)
			}
		}
	})

	// Обрабатываем также одиночные теги code (не в pre)
	doc.Find("code").Each(func(i int, s *goquery.Selection) {
		// Проверяем, что это не внутри pre (чтобы не обрабатывать дважды)
		if s.Parent().Is("pre") {
			return
		}

		// Извлекаем язык программирования
		language := extractLanguageFromCode(s)

		if language != "" {
			// Добавляем класс language-{lang} к тегу code
			currentClass := s.AttrOr("class", "")
			if !strings.Contains(currentClass, "language-"+language) {
				newClass := currentClass
				if newClass != "" {
					newClass += " "
				}
				newClass += "language-" + language
				s.SetAttr("class", newClass)
			}
		}
	})

	// Возвращаем обработанный HTML
	html, err := doc.Html()
	if err != nil {
		return htmlContent
	}

	return html
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

		// Erlang
		"erlang": "erlang",

		// PHP
		"php": "php",

		// Haskell
		"haskell": "haskell",
		"hs":      "haskell",

		// Scala
		"scala": "scala",

		// Clojure
		"clojure": "clojure",
		"clj":     "clojure",

		// Elixir
		"elixir": "elixir",
		"ex":     "elixir",

		// F#
		"f#":     "fsharp",
		"fsharp": "fsharp",

		// OCaml
		"ocaml": "ocaml",
		"ml":    "ocaml",

		// R
		"r": "r",

		// Julia
		"julia": "julia",
		"jl":    "julia",

		// Lua
		"lua": "lua",

		// Perl
		"perl": "perl",
		"pl":   "perl",

		// Groovy
		"groovy": "groovy",

		// Dart
		"dart": "dart",

		// Nim
		"nim": "nim",

		// Crystal
		"crystal": "crystal",
		"cr":      "crystal",

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

	// Если не нашли в маппинге, но это похоже на язык (короткое название без пробелов)
	if len(name) > 0 && len(name) <= 20 && !strings.Contains(name, " ") {
		return "go" // Fallback на язык C для неизвестных языков
	}

	return "go"
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
