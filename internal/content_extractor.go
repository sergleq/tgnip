package internal

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
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
)

// Content представляет извлеченный контент
type Content struct {
	Title    string
	Markdown string
	URL      string
	Author   string
	Date     string
}

// CollyConfig содержит конфигурацию для Colly
type CollyConfig struct {
	UserAgent      string
	Timeout        time.Duration
	MaxRetries     int
	FollowRedirect bool
	RespectRobots  bool
}

// DefaultCollyConfig возвращает конфигурацию по умолчанию
func DefaultCollyConfig() *CollyConfig {
	return &CollyConfig{
		UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		Timeout:        30 * time.Second,
		MaxRetries:     3,
		FollowRedirect: true,
		RespectRobots:  false,
	}
}

// createCollyCollector создает и настраивает коллектор Colly
func createCollyCollector(config *CollyConfig) *colly.Collector {
	c := colly.NewCollector(
		colly.UserAgent(config.UserAgent),
		colly.AllowURLRevisit(),
		colly.MaxDepth(1),
	)

	// Настройка таймаута
	c.SetRequestTimeout(config.Timeout)

	// Настройка лимитов
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		RandomDelay: 1 * time.Second,
	})

	// Добавляем расширения
	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	// Настройка обработки ошибок
	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Ошибка при загрузке %s: %v\n", r.Request.URL, err)
	})

	// Настройка обработки редиректов
	if config.FollowRedirect {
		// Colly автоматически обрабатывает редиректы, но мы можем настроить лимиты
		c.Limit(&colly.LimitRule{
			DomainGlob:  "*",
			Parallelism: 1,
			RandomDelay: 1 * time.Second,
			Delay:       2 * time.Second,
		})
	}

	return c
}

// ExtractContent извлекает контент из веб-страницы с использованием Colly
func ExtractContent(pageURL string) (*Content, error) {
	return ExtractContentWithConfig(pageURL, DefaultCollyConfig())
}

// ExtractContentWithConfig извлекает контент с пользовательской конфигурацией
func ExtractContentWithConfig(pageURL string, config *CollyConfig) (*Content, error) {
	fmt.Printf("Извлекаю контент из: %s\n", pageURL)

	// Создаем коллектор Colly
	c := createCollyCollector(config)

	// Переменные для хранения данных
	var htmlContent string
	var finalURL string
	var loadError error

	// Настраиваем обработчик для получения HTML
	c.OnResponse(func(r *colly.Response) {
		finalURL = r.Request.URL.String()
		htmlContent = string(r.Body)
		fmt.Printf("Успешно загружена страница: %s (размер: %d байт)\n", finalURL, len(htmlContent))
	})

	// Настраиваем обработчик ошибок
	c.OnError(func(r *colly.Response, err error) {
		loadError = fmt.Errorf("ошибка при загрузке %s: %w", r.Request.URL, err)
	})

	// Выполняем запрос
	err := c.Visit(pageURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка при посещении страницы: %w", err)
	}

	// Проверяем ошибки загрузки
	if loadError != nil {
		return nil, loadError
	}

	// Проверяем, что контент загружен
	if htmlContent == "" {
		return nil, fmt.Errorf("не удалось загрузить контент страницы")
	}

	// Парсим HTML с помощью goquery для извлечения метаданных
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("ошибка при парсинге HTML: %w", err)
	}

	// Парсим URL для go-readability
	parsedURL, err := url.Parse(finalURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка при парсинге URL: %w", err)
	}

	// Извлекаем контент с помощью go-readability
	article, err := readability.FromReader(strings.NewReader(htmlContent), parsedURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка при извлечении контента: %w", err)
	}

	// Создаем объект контента
	content := &Content{
		URL: finalURL,
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

// ExtractContentWithFallback извлекает контент с fallback на стандартный HTTP клиент
func ExtractContentWithFallback(pageURL string) (*Content, error) {
	// Сначала пробуем с Colly
	content, err := ExtractContent(pageURL)
	if err == nil {
		return content, nil
	}

	fmt.Printf("Colly не удалось загрузить страницу, пробуем fallback: %v\n", err)

	// Fallback на стандартный HTTP клиент
	return extractContentWithHTTPClient(pageURL)
}

// extractContentWithHTTPClient извлекает контент с помощью стандартного HTTP клиента
func extractContentWithHTTPClient(pageURL string) (*Content, error) {
	fmt.Printf("Использую fallback HTTP клиент для: %s\n", pageURL)

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

	fmt.Printf("Извлечено (fallback): заголовок='%s', длина markdown=%d символов\n",
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
func IsValidURL(str string) bool {
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
