package main

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
			result := isValidURL(tt.url)
			if result != tt.expected {
				t.Errorf("isValidURL(%s) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestExtractLanguageFromCode(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "Language prefix format",
			html:     `<code class="language-python">print("Hello")</code>`,
			expected: "python",
		},
		{
			name:     "Lang prefix format",
			html:     `<code class="lang-javascript">console.log("Hello")</code>`,
			expected: "javascript",
		},
		{
			name:     "Direct language name",
			html:     `<code class="go">fmt.Println("Hello")</code>`,
			expected: "go",
		},
		{
			name:     "Multiple classes",
			html:     `<code class="highlight language-cpp">int main()</code>`,
			expected: "cpp",
		},
		{
			name:     "No language class",
			html:     `<code>just text</code>`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			codeElement := doc.Find("code")
			result := extractLanguageFromCode(codeElement)

			if result != tt.expected {
				t.Errorf("extractLanguageFromCode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractCodeLanguagesInOrder(t *testing.T) {
	html := `
	<pre><code class="language-go">package main
func main() {
    fmt.Println("Hello")
}</code></pre>
	
	<pre><code class="javascript">function test() {
    console.log("test")
}</code></pre>
	
	<pre><code class="python">def test():
    print("test")</code></pre>
	`

	languages := extractCodeLanguagesInOrder(html)

	expectedLanguages := []string{"go", "javascript", "python"}
	if len(languages) != len(expectedLanguages) {
		t.Errorf("Expected %d languages, got %d", len(expectedLanguages), len(languages))
	}

	for i, expected := range expectedLanguages {
		if i < len(languages) && languages[i] != expected {
			t.Errorf("Expected language %s at position %d, got %s", expected, i, languages[i])
		}
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

	languages := extractCodeLanguagesInOrder(article.Content)
	result := extractAndConvertToMarkdown(article, languages)

	// Проверяем, что результат не пустой
	if result == "" {
		t.Error("Expected non-empty result")
	}

	// Проверяем, что языки были применены
	if !strings.Contains(result, "```go") {
		t.Error("Expected Go language to be applied")
	}

	if !strings.Contains(result, "```javascript") {
		t.Error("Expected JavaScript language to be applied")
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
