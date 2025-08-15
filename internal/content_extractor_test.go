package internal

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	readability "github.com/go-shiori/go-readability"
)

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"Valid HTTP URL", "http://example.com", true},
		{"Valid HTTPS URL", "https://example.com", true},
		{"Valid URL with path", "https://example.com/path", true},
		{"Invalid URL - no scheme", "example.com", false},
		{"Invalid URL - no host", "http://", false},
		{"Invalid URL - empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidURL(tt.url)
			if result != tt.expected {
				t.Errorf("isValidURL(%s) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestExtractAndConvertToMarkdown(t *testing.T) {
	html := `
	<article>
		<h1>Test Article</h1>
		<p>This is a test paragraph.</p>
		
		<pre><code class="language-go">package main
func main() {
    fmt.Println("Hello")
}</code></pre>
		
		<p>Another paragraph.</p>
		
		<pre><code class="javascript">function test() {
    console.log("test")
}</code></pre>
	</article>
	`

	article := readability.Article{
		Content: html,
	}

	result := extractAndConvertToMarkdown(article)

	// Проверяем, что результат не пустой
	if result == "" {
		t.Error("Expected non-empty result")
	}

	// Проверяем, что нет множественных пустых строк
	if strings.Contains(result, "\n\n\n") {
		t.Error("Result contains multiple consecutive empty lines")
	}
}

func TestConvertTableToMarkdown(t *testing.T) {
	html := `
	<table>
		<tr><th>Name</th><th>Age</th></tr>
		<tr><td>John</td><td>25</td></tr>
		<tr><td>Jane</td><td>30</td></tr>
	</table>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	table := doc.Find("table")
	result := convertTableToMarkdown(table)

	if result == nil {
		t.Error("Expected non-nil result")
		return
	}

	markdown := *result

	// Проверяем, что таблица была конвертирована
	if !strings.Contains(markdown, "|") {
		t.Error("Expected table to contain pipe separators")
	}

	if !strings.Contains(markdown, "Name") || !strings.Contains(markdown, "Age") {
		t.Error("Expected table headers to be present")
	}
}

func TestCleanTitleForFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple title",
			input:    "Hello World",
			expected: "Hello_World",
		},
		{
			name:     "Title with special characters",
			input:    "Hello: World! (Test)",
			expected: "Hello_World!_(Test)",
		},
		{
			name:     "Title with multiple spaces",
			input:    "Hello   World",
			expected: "Hello_World",
		},
		{
			name:     "Title with dots and commas",
			input:    "Hello, World. Test.",
			expected: "Hello,_World._Test.",
		},
		{
			name:     "Title with quotes",
			input:    `"Hello" 'World'`,
			expected: "Hello_World",
		},
		{
			name:     "Title with brackets",
			input:    "[Hello] {World}",
			expected: "[Hello]_{World}",
		},
		{
			name:     "Title with slashes",
			input:    "Hello/World\\Test",
			expected: "Hello_World_Test",
		},
		{
			name:     "Title with multiple underscores",
			input:    "Hello___World",
			expected: "Hello_World",
		},
		{
			name:     "Title with leading/trailing underscores",
			input:    "_Hello_World_",
			expected: "Hello_World",
		},
		{
			name:     "Title with Cyrillic characters",
			input:    "Привет Мир!",
			expected: "Привет_Мир!",
		},
		{
			name:     "Title with mixed languages",
			input:    "Hello Мир! (Test)",
			expected: "Hello_Мир!_(Test)",
		},
		{
			name:     "Title with accented characters",
			input:    "Café Español",
			expected: "Café_Español",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanTitleForFilename(tt.input)
			if result != tt.expected {
				t.Errorf("cleanTitleForFilename(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateFilename(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		title    string
		expected string
	}{
		{
			name:     "With valid title",
			url:      "https://example.com/article",
			title:    "Hello World",
			expected: "Hello_World.md",
		},
		{
			name:     "With empty title",
			url:      "https://example.com/article",
			title:    "",
			expected: "example_com_article.md",
		},
		{
			name:     "With 'Без заголовка' title",
			url:      "https://example.com/article",
			title:    "Без заголовка",
			expected: "example_com_article.md",
		},
		{
			name:     "With special characters in title",
			url:      "https://example.com/article",
			title:    "Hello: World! (Test)",
			expected: "Hello_World!_(Test).md",
		},
		{
			name:     "With Cyrillic title",
			url:      "https://example.com/article",
			title:    "Привет Мир!",
			expected: "Привет_Мир!.md",
		},
		{
			name:     "With mixed language title",
			url:      "https://example.com/article",
			title:    "Hello Мир! (Test)",
			expected: "Hello_Мир!_(Test).md",
		},
		{
			name:     "With invalid title (only invalid chars)",
			url:      "https://example.com/article",
			title:    "////",
			expected: "new_markdown_file.md",
		},
		{
			name:     "With invalid title (only underscores after cleaning)",
			url:      "https://example.com/article",
			title:    "////",
			expected: "new_markdown_file.md",
		},
		{
			name:     "With invalid title (starts with dot)",
			url:      "https://example.com/article",
			title:    ".hidden",
			expected: "new_markdown_file.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateFilename(tt.url, tt.title)
			if result != tt.expected {
				t.Errorf("generateFilename(%s, %s) = %v, want %v", tt.url, tt.title, result, tt.expected)
			}
		})
	}
}

func TestIsValidFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "Valid filename",
			filename: "hello_world.md",
			expected: true,
		},
		{
			name:     "Valid filename with Cyrillic",
			filename: "привет_мир.md",
			expected: true,
		},
		{
			name:     "Valid filename with special chars",
			filename: "hello-world!_(test).md",
			expected: true,
		},
		{
			name:     "Empty filename",
			filename: "",
			expected: false,
		},
		{
			name:     "Only extension",
			filename: ".md",
			expected: false,
		},
		{
			name:     "Too long filename",
			filename: strings.Repeat("a", 256) + ".md",
			expected: false,
		},
		{
			name:     "Hidden file",
			filename: ".hidden.md",
			expected: false,
		},
		{
			name:     "Ends with dot before extension",
			filename: "filename..md",
			expected: false,
		},
		{
			name:     "Only underscores",
			filename: "____.md",
			expected: false,
		},
		{
			name:     "Only underscore",
			filename: "_.md",
			expected: false,
		},
		{
			name:     "Contains underscores but valid",
			filename: "hello_world.md",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidFilename(tt.filename)
			if result != tt.expected {
				t.Errorf("isValidFilename(%s) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestConvertToMarkdownLocalization(t *testing.T) {
	content := &Content{
		Title:    "Test Article",
		Markdown: "This is test content.",
		Author:   "John Doe",
		Date:     "2025-01-01",
		URL:      "https://example.com/article",
	}

	tests := []struct {
		name     string
		locale   Locale
		expected []string // Строки, которые должны присутствовать в markdown
	}{
		{
			name: "Russian localization",
			locale: Locale{
				MetadataSection: "## Метаинформация",
				SourceLabel:     "**Источник:**",
				AuthorLabel:     "**Автор:**",
				DateLabel:       "**Дата:**",
				ProcessedLabel:  "**Обработано:**",
				ContentSection:  "## Содержание",
				FooterText:      "*Этот документ был автоматически создан ботом для преобразования веб-страниц в markdown формат.*",
				UnknownSource:   "Неизвестный источник",
			},
			expected: []string{
				"# Test Article",
				"## Метаинформация",
				"**Источник:**",
				"**Автор:**",
				"**Дата:**",
				"**Обработано:**",
				"## Содержание",
				"*Этот документ был автоматически создан ботом для преобразования веб-страниц в markdown формат.*",
			},
		},
		{
			name: "English localization",
			locale: Locale{
				MetadataSection: "## Metadata",
				SourceLabel:     "**Source:**",
				AuthorLabel:     "**Author:**",
				DateLabel:       "**Date:**",
				ProcessedLabel:  "**Processed:**",
				ContentSection:  "## Content",
				FooterText:      "*This document was automatically generated by a bot for converting web pages to markdown format.*",
				UnknownSource:   "Unknown source",
			},
			expected: []string{
				"# Test Article",
				"## Metadata",
				"**Source:**",
				"**Author:**",
				"**Date:**",
				"**Processed:**",
				"## Content",
				"*This document was automatically generated by a bot for converting web pages to markdown format.*",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertToMarkdown(content, "https://example.com/article", tt.locale)

			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected markdown to contain '%s', but it doesn't", expected)
				}
			}
		})
	}
}
