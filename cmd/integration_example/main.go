package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/src-d/enry/v2"
)

// Content представляет извлеченный контент
type Content struct {
	Title    string
	Markdown string
	URL      string
	Author   string
	Date     string
}

// Locale представляет локализацию для определенного языка
type Locale struct {
	WelcomeMessage        string
	ProcessingMessage     string
	InvalidURLMessage     string
	ErrorProcessingMsg    string
	ErrorSendingMsg       string
	SuccessMessage        string
	ServerOverloadMessage string

	// Локализация для markdown файлов
	MetadataSection string
	SourceLabel     string
	AuthorLabel     string
	DateLabel       string
	ProcessedLabel  string
	ContentSection  string
	FooterText      string
	UnknownSource   string
}

// LanguageScore представляет оценку для конкретного языка
type LanguageScore struct {
	Language   string
	Score      float64
	Confidence float64
}

// CodeBlock представляет блок кода с контекстом
type CodeBlock struct {
	Content     string
	Language    string
	Context     string
	LineNumber  int
	HasExplicit bool
}

// LanguageDetector представляет детектор языка программирования
type LanguageDetector struct {
	contextWeight     float64
	fileExtWeight     float64
	commandWeight     float64
	heuristicWeight   float64
	enryWeight        float64
	explicitWeight    float64
	contextWindowSize int
}

// locales содержит все поддерживаемые языки
var locales = map[string]Locale{
	"ru": {
		WelcomeMessage:        "Привет! Я бот для преобразования ссылок в markdown файлы.",
		ProcessingMessage:     "⏳ Обрабатываю ссылку...",
		InvalidURLMessage:     "Пожалуйста, отправьте валидную ссылку на веб-страницу.",
		ErrorProcessingMsg:    "❌ Не удалось обработать ссылку.",
		ErrorSendingMsg:       "❌ Не удалось отправить файл.",
		SuccessMessage:        "✅ Файл успешно создан!",
		ServerOverloadMessage: "⚠️ Сервер перегружен. Попробуйте позже.",
		MetadataSection:       "## Метаинформация",
		SourceLabel:           "**Источник:**",
		AuthorLabel:           "**Автор:**",
		DateLabel:             "**Дата:**",
		ProcessedLabel:        "**Обработано:**",
		ContentSection:        "## Содержание",
		FooterText:            "*Этот документ был автоматически создан ботом.*",
		UnknownSource:         "Неизвестный источник",
	},
}

// NewLanguageDetector создает новый детектор языка
func NewLanguageDetector() *LanguageDetector {
	return &LanguageDetector{
		contextWeight:     3.0,
		fileExtWeight:     3.0,
		commandWeight:     2.0,
		heuristicWeight:   2.0,
		enryWeight:        2.0,
		explicitWeight:    10.0,
		contextWindowSize: 3,
	}
}

// DetectLanguage определяет язык для одного блока кода
func (ld *LanguageDetector) DetectLanguage(block CodeBlock) LanguageScore {
	// Если есть явное указание языка, используем его
	if block.HasExplicit {
		return LanguageScore{
			Language:   block.Language,
			Score:      ld.explicitWeight,
			Confidence: 1.0,
		}
	}

	// Собираем все сигналы
	scores := make(map[string]float64)

	// Эвристики по коду
	ld.addCodeHeuristicSignals(block.Content, scores)

	// Результат внешнего детектора (enry)
	ld.addEnrySignals(block.Content, scores)

	// Находим язык с максимальным скором
	var bestLanguage string
	var maxScore float64

	for lang, score := range scores {
		if score > maxScore {
			maxScore = score
			bestLanguage = lang
		}
	}

	// Вычисляем confidence
	totalScore := 0.0
	for _, score := range scores {
		totalScore += score
	}

	confidence := 0.0
	if totalScore > 0 {
		confidence = maxScore / totalScore
	}

	return LanguageScore{
		Language:   bestLanguage,
		Score:      maxScore,
		Confidence: confidence,
	}
}

// addCodeHeuristicSignals добавляет эвристические сигналы из кода
func (ld *LanguageDetector) addCodeHeuristicSignals(content string, scores map[string]float64) {
	// Go сигнатуры
	if strings.Contains(content, "package main") ||
		(strings.Contains(content, "func ") && !strings.Contains(content, "def ")) ||
		strings.Contains(content, "import \"") ||
		strings.Contains(content, ":=") ||
		strings.Contains(content, "int64") ||
		strings.Contains(content, "fmt.") {
		scores["go"] += ld.heuristicWeight
	}

	// Python сигнатуры
	if strings.Contains(content, "def ") ||
		strings.Contains(content, "import ") ||
		strings.Contains(content, "self.") ||
		strings.Contains(content, "if __name__") ||
		!strings.Contains(content, ";") && strings.Contains(content, "print(") {
		scores["python"] += ld.heuristicWeight
	}

	// JavaScript сигнатуры
	if (strings.Contains(content, "import ") && !strings.Contains(content, "import \"")) ||
		strings.Contains(content, "export ") ||
		strings.Contains(content, "async ") ||
		strings.Contains(content, "await ") ||
		strings.Contains(content, "const ") ||
		strings.Contains(content, "let ") {
		scores["javascript"] += ld.heuristicWeight
	}

	// SQL сигнатуры
	if strings.Contains(content, "SELECT ") ||
		strings.Contains(content, "FROM ") ||
		strings.Contains(content, "WHERE ") ||
		strings.Contains(content, "JOIN ") ||
		strings.Contains(content, "INSERT INTO") {
		scores["sql"] += ld.heuristicWeight
	}

	// Bash сигнатуры
	if strings.Contains(content, "#!/bin/bash") ||
		strings.Contains(content, "#!/bin/sh") ||
		strings.Contains(content, "cd ") ||
		strings.Contains(content, "grep ") ||
		strings.Contains(content, "sed ") ||
		strings.Contains(content, "$") {
		scores["bash"] += ld.heuristicWeight
	}
}

// addEnrySignals добавляет сигналы от внешнего детектора enry
func (ld *LanguageDetector) addEnrySignals(content string, scores map[string]float64) {
	// Используем enry для детекции языка
	detectedLang := enry.GetLanguage("", []byte(content))

	if detectedLang != "" {
		// Маппинг языков enry к нашим языкам
		langMapping := map[string]string{
			"Go":         "go",
			"Python":     "python",
			"JavaScript": "javascript",
			"TypeScript": "typescript",
			"Rust":       "rust",
			"Java":       "java",
			"C++":        "c++",
			"C":          "c",
			"Shell":      "bash",
			"SQL":        "sql",
			"YAML":       "yaml",
			"TOML":       "toml",
			"JSON":       "json",
		}

		if mappedLang, ok := langMapping[detectedLang]; ok {
			// Вес зависит от длины сниппета
			weight := ld.enryWeight
			if len(content) < 50 {
				weight *= 0.5
			} else if len(content) > 200 {
				weight *= 1.5
			}
			scores[mappedLang] += weight
		}
	}
}

// processMarkdownWithLanguageDetection обрабатывает Markdown и добавляет языки к блокам кода
func processMarkdownWithLanguageDetection(markdown string) string {
	// Создаем детектор языка
	detector := NewLanguageDetector()

	// Находим все блоки кода
	codeBlockRegex := regexp.MustCompile(`(?s)` + "```" + `(\w*)\s*\n(.*?)\n` + "```" + ``)
	matches := codeBlockRegex.FindAllStringSubmatch(markdown, -1)

	// Если блоков кода нет, возвращаем исходный текст
	if len(matches) == 0 {
		return markdown
	}

	// Обрабатываем каждый блок кода
	result := markdown
	for _, match := range matches {
		if len(match) >= 3 {
			originalBlock := match[0]
			language := match[1]
			content := match[2]

			// Если язык уже указан, пропускаем
			if language != "" {
				continue
			}

			// Создаем блок кода для анализа
			block := CodeBlock{
				Content:     content,
				Language:    "",
				Context:     "",
				LineNumber:  0,
				HasExplicit: false,
			}

			// Определяем язык
			score := detector.DetectLanguage(block)

			// Если уверенность достаточно высокая, добавляем язык
			if score.Confidence > 0.3 && score.Language != "" {
				// Создаем новый блок с языком
				newBlock := "```" + score.Language + "\n" + content + "\n```"

				// Заменяем в результате
				result = strings.Replace(result, originalBlock, newBlock, 1)
			}
		}
	}

	return result
}

// convertToMarkdown конвертирует контент в markdown формат
func convertToMarkdown(content *Content, originalURL string, locale Locale) string {
	var markdown strings.Builder

	// Заголовок
	markdown.WriteString(fmt.Sprintf("# %s\n\n", content.Title))

	// Метаинформация
	markdown.WriteString(fmt.Sprintf("%s\n\n", locale.MetadataSection))
	markdown.WriteString(fmt.Sprintf("- %s [%s](%s)\n", locale.SourceLabel, "example.com", originalURL))

	if content.Author != "" {
		markdown.WriteString(fmt.Sprintf("- %s %s\n", locale.AuthorLabel, content.Author))
	}

	if content.Date != "" {
		markdown.WriteString(fmt.Sprintf("- %s %s\n", locale.DateLabel, content.Date))
	}

	markdown.WriteString("\n---\n\n")

	// Основной текст
	markdown.WriteString(fmt.Sprintf("%s\n\n", locale.ContentSection))

	// Обрабатываем контент с детекцией языка для блоков кода
	processedContent := processMarkdownWithLanguageDetection(content.Markdown)
	markdown.WriteString(processedContent)
	markdown.WriteString("\n\n")

	// Футер
	markdown.WriteString("---\n\n")
	markdown.WriteString(fmt.Sprintf("%s\n", locale.FooterText))

	return markdown.String()
}

func main() {
	// Пример контента с блоками кода без указания языка
	content := &Content{
		Title:    "Пример статьи с кодом",
		Markdown: "# Программирование на разных языках\n\n## Go пример\n\nВ Go вы можете использовать горутины для конкурентности:\n\n```\npackage main\n\nimport (\n    \"fmt\"\n    \"time\"\n)\n\nfunc main() {\n    go func() {\n        fmt.Println(\"Горутина работает\")\n    }()\n    time.Sleep(time.Second)\n}\n```\n\n## Python пример\n\nВ Python с Django:\n\n```\nfrom django.shortcuts import render\nfrom django.http import JsonResponse\n\ndef api_view(request):\n    data = {\"message\": \"Hello from Django!\"}\n    return JsonResponse(data)\n```\n\nЗапустите с: python manage.py runserver\n\n## JavaScript пример\n\nСоздайте файл main.js:\n\n```\nconst express = require('express');\nconst app = express();\n\napp.get('/', (req, res) => {\n    res.json({ message: 'Hello from Express!' });\n});\n\napp.listen(3000, () => {\n    console.log('Server running on port 3000');\n});\n```\n\nИ запустите: node main.js\n\n## SQL запрос\n\n```\nSELECT u.name, u.email, COUNT(p.id) as post_count\nFROM users u\nLEFT JOIN posts p ON u.id = p.user_id\nWHERE u.active = true\nGROUP BY u.id, u.name, u.email\nORDER BY post_count DESC;\n```\n\n## Блок без указания языка\n\n```\n#!/bin/bash\necho \"Проверяем статус сервиса...\"\nsystemctl status nginx\nif [ $? -eq 0 ]; then\n    echo \"Nginx работает\"\nelse\n    echo \"Nginx не запущен\"\nfi\n```",
		Author:   "Автор статьи",
		Date:     "2024-01-01",
		URL:      "https://example.com/article",
	}

	// Получаем русскую локализацию
	locale := locales["ru"]

	// Конвертируем в Markdown с автоматической детекцией языка
	result := convertToMarkdown(content, "https://example.com/article", locale)

	fmt.Println("=== РЕЗУЛЬТАТ КОНВЕРТАЦИИ ===\n")
	fmt.Println(result)

	// Анализируем результат
	fmt.Println("\n=== АНАЛИЗ РЕЗУЛЬТАТА ===")

	// Проверяем, какие языки были детектированы
	languages := []string{"go", "python", "javascript", "sql", "bash"}
	for _, lang := range languages {
		if strings.Contains(result, "```"+lang+"\n") {
			fmt.Printf("✅ Язык %s детектирован и добавлен\n", lang)
		} else {
			fmt.Printf("❌ Язык %s НЕ детектирован\n", lang)
		}
	}

	// Проверяем сохранение контента
	if strings.Contains(result, "Программирование на разных языках") {
		fmt.Println("✅ Заголовок сохранен")
	}

	if strings.Contains(result, "Автор статьи") {
		fmt.Println("✅ Автор сохранен")
	}

	if strings.Contains(result, "Горутина работает") {
		fmt.Println("✅ Содержимое кода сохранено")
	}
}
