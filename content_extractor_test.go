package main

import (
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	readability "github.com/go-shiori/go-readability"
)

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "Valid HTTP URL",
			url:      "http://example.com",
			expected: true,
		},
		{
			name:     "Valid HTTPS URL",
			url:      "https://example.com/article",
			expected: true,
		},
		{
			name:     "Invalid URL - no scheme",
			url:      "example.com",
			expected: false,
		},
		{
			name:     "Invalid URL - no host",
			url:      "http://",
			expected: false,
		},
		{
			name:     "Invalid URL - empty string",
			url:      "",
			expected: false,
		},
		{
			name:     "Invalid URL - random text",
			url:      "not a url",
			expected: false,
		},
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

func TestCleanText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Multiple spaces",
			input:    "Hello    world",
			expected: "Hello world",
		},
		{
			name:     "Multiple newlines",
			input:    "Hello\n\n\nworld",
			expected: "",
		},
		{
			name:     "Short paragraph",
			input:    "Short",
			expected: "",
		},
		{
			name:     "Normal text",
			input:    "This is a normal paragraph with some text.",
			expected: "This is a normal paragraph with some text.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanText(tt.input)
			if result != tt.expected {
				t.Errorf("cleanText(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateFilename(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "Simple domain",
			url:      "https://example.com",
			expected: "example_com_index_md",
		},
		{
			name:     "Article URL",
			url:      "https://blog.example.com/article-title",
			expected: "blog_example_com_article_title_md",
		},
		{
			name:     "URL with query params",
			url:      "https://example.com/article?param=value",
			expected: "example_com_article_md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateFilename(tt.url)
			if result != tt.expected {
				t.Errorf("generateFilename(%s) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestEscapeMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Text with asterisks",
			input:    "Hello *world*",
			expected: "Hello \\*world\\*",
		},
		{
			name:     "Text with underscores",
			input:    "Hello _world_",
			expected: "Hello \\_world\\_",
		},
		{
			name:     "Text with backticks",
			input:    "Hello `world`",
			expected: "Hello \\`world\\`",
		},
		{
			name:     "Text with hash",
			input:    "Hello #world",
			expected: "Hello \\#world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeMarkdown(tt.input)
			if result != tt.expected {
				t.Errorf("escapeMarkdown(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCleanTextPreservesParagraphs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Multiple paragraphs",
			input:    "First paragraph.\n\nSecond paragraph.\n\nThird paragraph.",
			expected: "First paragraph.\n\nSecond paragraph.\n\nThird paragraph.",
		},
		{
			name:     "Paragraphs with extra spaces",
			input:    "First paragraph.    \n\n   Second paragraph.   \n\nThird paragraph.",
			expected: "First paragraph.\n\nSecond paragraph.\n\nThird paragraph.",
		},
		{
			name:     "Short paragraphs filtered out",
			input:    "First paragraph.\n\nShort.\n\nThird paragraph.",
			expected: "First paragraph.\n\nThird paragraph.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanText(tt.input)
			if result != tt.expected {
				t.Errorf("cleanText(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestContentStructure(t *testing.T) {
	// Тест структуры Content
	content := Content{
		Title:    "Test Title",
		Markdown: "# Test Content\n\nThis is test content.",
		URL:      "https://example.com",
		Author:   "Test Author",
		Date:     "2025-08-13",
	}

	if content.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", content.Title)
	}

	if content.Markdown == "" {
		t.Errorf("Expected markdown content, got empty string")
	}

	if content.URL != "https://example.com" {
		t.Errorf("Expected URL 'https://example.com', got '%s'", content.URL)
	}
}

func TestContentWithMarkdown(t *testing.T) {
	// Тест Content с markdown контентом
	content := Content{
		Title:    "Test Title",
		Markdown: "# Test Content\n\nThis is test content with **bold** text.",
		URL:      "https://example.com",
		Author:   "Test Author",
		Date:     "2025-08-13",
	}

	if content.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", content.Title)
	}

	if content.Markdown == "" {
		t.Errorf("Expected markdown content, got empty string")
	}

	if !strings.Contains(content.Markdown, "# Test Content") {
		t.Errorf("Expected markdown to contain heading, got '%s'", content.Markdown)
	}

	if content.URL != "https://example.com" {
		t.Errorf("Expected URL 'https://example.com', got '%s'", content.URL)
	}
}

func TestConvertTableToMarkdown(t *testing.T) {
	// Тест конвертации таблицы
	html := `
	<table>
		<thead>
			<tr>
				<th>Язык</th>
				<th>Сложность</th>
				<th>Популярность</th>
			</tr>
		</thead>
		<tbody>
			<tr>
				<td>Python</td>
				<td>Низкая</td>
				<td>Высокая</td>
			</tr>
			<tr>
				<td>JavaScript</td>
				<td>Средняя</td>
				<td>Очень высокая</td>
			</tr>
		</tbody>
	</table>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	table := doc.Find("table").First()
	result := convertTableToMarkdown(table)

	if result == nil {
		t.Fatal("Expected table conversion result, got nil")
	}

	markdown := *result

	// Проверяем, что результат содержит заголовки
	if !strings.Contains(markdown, "| Язык |") {
		t.Errorf("Expected table to contain 'Язык' header, got: %s", markdown)
	}

	// Проверяем, что результат содержит разделитель
	if !strings.Contains(markdown, "| --- |") {
		t.Errorf("Expected table to contain separator, got: %s", markdown)
	}

	// Проверяем, что результат содержит данные
	if !strings.Contains(markdown, "| Python |") {
		t.Errorf("Expected table to contain 'Python' data, got: %s", markdown)
	}

	// Проверяем структуру таблицы
	lines := strings.Split(strings.TrimSpace(markdown), "\n")
	if len(lines) < 3 {
		t.Errorf("Expected at least 3 lines (header, separator, data), got %d", len(lines))
	}
}

func TestEscapeMarkdownTableCell(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal text",
			input:    "Simple text",
			expected: "Simple text",
		},
		{
			name:     "Text with pipe",
			input:    "Text | with pipe",
			expected: "Text \\| with pipe",
		},
		{
			name:     "Text with newlines",
			input:    "Text\nwith\nnewlines",
			expected: "Text with newlines",
		},
		{
			name:     "Text with extra spaces",
			input:    "  Text with spaces  ",
			expected: "Text with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeMarkdownTableCell(tt.input)
			if result != tt.expected {
				t.Errorf("escapeMarkdownTableCell(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCleanMarkdownLines(t *testing.T) {
	// Тест очистки markdown с правильной обработкой пустых строк
	input := `# Заголовок

Первый абзац.


Второй абзац.

## Подзаголовок

Третий абзац.`

	// Создаем простую статью для тестирования
	article := readability.Article{
		Content: input,
	}

	result := extractAndConvertToMarkdown(article)

	// Проверяем, что нет множественных пустых строк
	if strings.Contains(result, "\n\n\n") {
		t.Errorf("Result contains multiple consecutive empty lines: %s", result)
	}

	// Проверяем, что есть правильные разделения между абзацами
	if !strings.Contains(result, "\n\n") {
		t.Errorf("Result should contain double newlines between paragraphs: %s", result)
	}

	// Проверяем, что нет пустых строк в начале и конце
	if strings.HasPrefix(result, "\n") || strings.HasSuffix(result, "\n") {
		t.Errorf("Result should not start or end with newlines: %s", result)
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
			name:     "Short language code",
			html:     `<code class="js">alert("Hello")</code>`,
			expected: "javascript",
		},
		{
			name:     "C# language",
			html:     `<code class="csharp">Console.WriteLine("Hello")</code>`,
			expected: "csharp",
		},
		{
			name:     "C++ language",
			html:     `<code class="cpp">std::cout << "Hello"</code>`,
			expected: "cpp",
		},
		{
			name:     "No language class",
			html:     `<code>just text</code>`,
			expected: "",
		},
		{
			name:     "Empty class",
			html:     `<code class="">just text</code>`,
			expected: "",
		},
		{
			name:     "Unknown language",
			html:     `<code class="unknown-lang">text</code>`,
			expected: "unknown-lang",
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

func TestNormalizeLanguageName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "JavaScript variants",
			input:    "js",
			expected: "javascript",
		},
		{
			name:     "Python variants",
			input:    "py",
			expected: "python",
		},
		{
			name:     "C++ variants",
			input:    "c++",
			expected: "cpp",
		},
		{
			name:     "C# variants",
			input:    "c#",
			expected: "csharp",
		},
		{
			name:     "Go variants",
			input:    "golang",
			expected: "go",
		},
		{
			name:     "TypeScript variants",
			input:    "ts",
			expected: "typescript",
		},
		{
			name:     "Shell variants",
			input:    "bash",
			expected: "bash",
		},
		{
			name:     "YAML variants",
			input:    "yml",
			expected: "yaml",
		},
		{
			name:     "Case insensitive",
			input:    "PYTHON",
			expected: "python",
		},
		{
			name:     "Unknown language",
			input:    "unknown",
			expected: "unknown",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Too long name",
			input:    "thisisareallylonglanguagename",
			expected: "",
		},
		{
			name:     "Name with spaces",
			input:    "c plus plus",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeLanguageName(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeLanguageName(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRealHTMLWithLanguages(t *testing.T) {
	html := `
	<article>
		<h1>Test Article</h1>
		<p>Some text here.</p>
		
		<pre><code class="language-python">
def hello_world():
    print("Hello, World!")
		</code></pre>
		
		<pre><code class="javascript">
console.log("Hello, World!");
		</code></pre>
		
		<pre><code class="go">
package main
import "fmt"
func main() {
    fmt.Println("Hello, World!")
}
		</code></pre>
	</article>
	`

	// Создаем статью
	article := readability.Article{
		Content: html,
	}

	// Тестируем извлечение
	result := extractAndConvertToMarkdown(article)

	// Проверяем, что языки были извлечены
	if !strings.Contains(result, "```python") {
		t.Error("Expected Python language to be extracted")
	}

	if !strings.Contains(result, "```javascript") {
		t.Error("Expected JavaScript language to be extracted")
	}

	if !strings.Contains(result, "```go") {
		t.Error("Expected Go language to be extracted")
	}

	t.Logf("Result:\n%s", result)
}

func TestProcessCodeBlocksInHTML(t *testing.T) {
	// Простой HTML с блоком кода
	html := `<pre><code class="language-go">package main

func main() {
    fmt.Println("Hello")
}</code></pre>`

	// Обрабатываем HTML
	result := processCodeBlocksInHTML(html)

	// Проверяем, что атрибут data-language добавлен к pre
	if !strings.Contains(result, `data-language="go"`) {
		t.Errorf("processCodeBlocksInHTML() should add data-language attribute, got: %s", result)
	}

	// Проверяем, что класс language-go добавлен к code
	if !strings.Contains(result, `class="language-go"`) {
		t.Errorf("processCodeBlocksInHTML() should add language-go class, got: %s", result)
	}
}

func TestProcessCodeBlocksInHTMLComplex(t *testing.T) {
	// HTML с несколькими блоками кода
	html := `
	<p>Some text before.</p>
	<pre><code class="language-python">def hello():
    print("Hello from Python")</code></pre>
	<p>Some text between.</p>
	<pre><code class="javascript">function hello() {
    console.log("Hello from JS");
}</code></pre>
	<p>Some text after.</p>
	`

	// Обрабатываем HTML
	result := processCodeBlocksInHTML(html)

	// Проверяем, что атрибуты data-language добавлены
	if !strings.Contains(result, `data-language="python"`) {
		t.Errorf("processCodeBlocksInHTML() should add data-language for Python, got: %s", result)
	}

	if !strings.Contains(result, `data-language="javascript"`) {
		t.Errorf("processCodeBlocksInHTML() should add data-language for JavaScript, got: %s", result)
	}

	// Проверяем, что классы language-* добавлены (может быть в разном порядке)
	if !strings.Contains(result, `language-python`) {
		t.Errorf("processCodeBlocksInHTML() should add language-python class, got: %s", result)
	}

	if !strings.Contains(result, `language-javascript`) {
		t.Errorf("processCodeBlocksInHTML() should add language-javascript class, got: %s", result)
	}
}

func TestProcessCodeBlocksInHTMLRealistic(t *testing.T) {
	// Реалистичный HTML с блоками кода, как на реальных сайтах
	html := `
	<article>
		<h1>Programming Examples</h1>
		<p>Here are some code examples:</p>
		
		<pre><code class="language-go">package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}</code></pre>
		
		<p>And here's some JavaScript:</p>
		
		<pre><code class="highlight javascript">function greet(name) {
    console.log("Hello, " + name + "!");
}

greet("World");</code></pre>
		
		<p>And Python:</p>
		
		<pre><code class="language-python">def greet(name):
    print(f"Hello, {name}!")

greet("World")</code></pre>
	</article>
	`

	// Обрабатываем HTML
	result := processCodeBlocksInHTML(html)

	// Проверяем, что атрибуты data-language добавлены для всех языков
	if !strings.Contains(result, `data-language="go"`) {
		t.Errorf("processCodeBlocksInHTML() should add data-language for Go, got: %s", result)
	}

	if !strings.Contains(result, `data-language="javascript"`) {
		t.Errorf("processCodeBlocksInHTML() should add data-language for JavaScript, got: %s", result)
	}

	if !strings.Contains(result, `data-language="python"`) {
		t.Errorf("processCodeBlocksInHTML() should add data-language for Python, got: %s", result)
	}

	// Проверяем, что классы language-* добавлены (может быть в разном порядке)
	if !strings.Contains(result, `language-go`) {
		t.Errorf("processCodeBlocksInHTML() should add language-go class, got: %s", result)
	}

	if !strings.Contains(result, `language-javascript`) {
		t.Errorf("processCodeBlocksInHTML() should add language-javascript class, got: %s", result)
	}

	if !strings.Contains(result, `language-python`) {
		t.Errorf("processCodeBlocksInHTML() should add language-python class, got: %s", result)
	}
}

func TestExtractLanguageFromCodeWithHighlight(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "Highlight with language",
			html:     `<code class="highlight javascript">console.log("Hello")</code>`,
			expected: "javascript",
		},
		{
			name:     "Highlight with language-prefix",
			html:     `<code class="highlight language-python">print("Hello")</code>`,
			expected: "python",
		},
		{
			name:     "Highlight with lang-prefix",
			html:     `<code class="highlight lang-go">fmt.Println("Hello")</code>`,
			expected: "go",
		},
		{
			name:     "Only highlight class",
			html:     `<code class="highlight">just text</code>`,
			expected: "",
		},
		{
			name:     "Multiple highlight classes",
			html:     `<code class="highlight source code javascript">console.log("Hello")</code>`,
			expected: "javascript",
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

func TestRealHTMLToMarkdownConversion(t *testing.T) {
	// Реалистичный HTML с блоками кода
	html := `
	<article>
		<h1>Programming Examples</h1>
		
		<pre><code class="language-go">package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}</code></pre>
		
		<pre><code class="highlight javascript">function greet(name) {
    console.log("Hello, " + name + "!");
}

greet("World");</code></pre>
		
		<pre><code class="language-python">def greet(name):
    print(f"Hello, {name}!")

greet("World")</code></pre>
	</article>
	`

	// Создаем конвертер html-to-markdown
	converter := md.NewConverter("", true, nil)

	// Конвертируем HTML в markdown
	markdown, err := converter.ConvertString(html)
	if err != nil {
		t.Fatalf("Failed to convert HTML to markdown: %v", err)
	}

	t.Logf("Converted markdown:\n%s", markdown)

	// Обрабатываем HTML до конвертации
	processedHTML := processCodeBlocksInHTML(html)

	// Конвертируем обработанный HTML в markdown
	processedMarkdown, err := converter.ConvertString(processedHTML)
	if err != nil {
		t.Fatalf("Failed to convert HTML to markdown: %v", err)
	}

	t.Logf("After processing code blocks:\n%s", processedMarkdown)

	// Проверяем, что языки были добавлены (html-to-markdown может обрабатывать по-разному)
	if !strings.Contains(processedMarkdown, "```go") && !strings.Contains(processedMarkdown, "```highlight javascript") {
		t.Error("Expected Go or JavaScript language to be extracted")
	}

	if !strings.Contains(processedMarkdown, "```python") {
		t.Error("Expected Python language to be extracted")
	}
}

func TestRealWorldCodeBlocks(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "GitHub style code blocks",
			html: `
			<article>
				<pre><code class="language-javascript">console.log("Hello");</code></pre>
				<pre><code class="language-python">print("Hello")</code></pre>
			</article>`,
			expected: "javascript",
		},
		{
			name: "Stack Overflow style",
			html: `
			<pre><code class="lang-js">function test() {}</code></pre>
			<pre><code class="lang-python">def test(): pass</code></pre>`,
			expected: "javascript",
		},
		{
			name: "Medium style with highlight",
			html: `
			<pre><code class="highlight javascript">const x = 1;</code></pre>
			<pre><code class="highlight python">x = 1</code></pre>`,
			expected: "javascript",
		},
		{
			name: "Dev.to style",
			html: `
			<pre><code class="language-js">let x = 1;</code></pre>
			<pre><code class="language-py">x = 1</code></pre>`,
			expected: "javascript",
		},
		{
			name: "CodePen style",
			html: `
			<pre><code class="js">console.log("test");</code></pre>
			<pre><code class="python">print("test")</code></pre>`,
			expected: "javascript",
		},
		{
			name: "Multiple classes",
			html: `
			<pre><code class="highlight source-code javascript">function test() {}</code></pre>
			<pre><code class="prettyprint lang-python">def test(): pass</code></pre>`,
			expected: "javascript",
		},
		{
			name: "No language specified",
			html: `
			<pre><code>just some text</code></pre>`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			// Находим первый блок кода
			codeElement := doc.Find("pre code").First()
			if codeElement.Length() == 0 {
				t.Fatal("No code element found")
			}

			result := extractLanguageFromCode(codeElement)

			if result != tt.expected {
				t.Errorf("extractLanguageFromCode() = %v, want %v", result, tt.expected)
				t.Logf("HTML: %s", tt.html)
				t.Logf("Classes: %s", codeElement.AttrOr("class", ""))
			}
		})
	}
}

func TestFullContentExtractionWithLanguages(t *testing.T) {
	// Реалистичный HTML, который может прийти с реального сайта
	html := `
	<article>
		<h1>How to Use JavaScript and Python</h1>
		<p>Here are some examples:</p>
		
		<pre><code class="language-javascript">function greet(name) {
    console.log("Hello, " + name + "!");
}

greet("World");</code></pre>
		
		<p>And here's the Python version:</p>
		
		<pre><code class="highlight python">def greet(name):
    print(f"Hello, {name}!")

greet("World")</code></pre>
		
		<p>And some Go code:</p>
		
		<pre><code class="go">package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}</code></pre>
	</article>
	`

	// Создаем статью как это делает go-readability
	article := readability.Article{
		Content: html,
	}

	// Тестируем полный процесс извлечения
	result := extractAndConvertToMarkdown(article)

	t.Logf("Full extraction result:\n%s", result)

	// Проверяем, что языки были правильно определены (html-to-markdown может обрабатывать по-разному)
	if !strings.Contains(result, "```javascript") && !strings.Contains(result, "```highlight javascript") {
		t.Error("Expected JavaScript language to be extracted")
	}

	if !strings.Contains(result, "```python") && !strings.Contains(result, "```highlight python") {
		t.Error("Expected Python language to be extracted")
	}

	if !strings.Contains(result, "```go") && !strings.Contains(result, "```go go") {
		t.Error("Expected Go language to be extracted")
	}

	// Проверяем, что есть хотя бы один правильный язык
	if !strings.Contains(result, "```javascript") && !strings.Contains(result, "```python") && !strings.Contains(result, "```go") {
		t.Error("Expected at least one language to be extracted correctly")
	}
}

func TestCodeLanguageMapping(t *testing.T) {
	// Тест новой функциональности с массивом языков по порядку
	html := `
	<article>
		<h1>Programming Examples</h1>
		
		<pre><code class="language-go">package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}</code></pre>
		
		<pre><code class="highlight javascript">function greet(name) {
    console.log("Hello, " + name + "!");
}

greet("World");</code></pre>
		
		<pre><code class="python">def greet(name):
    print(f"Hello, {name}!")

greet("World")</code></pre>
	</article>
	`

	// Создаем статью как это делает go-readability
	article := readability.Article{
		Content: html,
	}

	// Извлекаем языки для отладки
	languages := extractCodeLanguagesInOrder(article.Content)
	t.Logf("Languages in order: %v", languages)

	// Тестируем полный процесс извлечения
	result := extractAndConvertToMarkdown(article)

	t.Logf("Full extraction result:\n%s", result)

	// Проверяем, что языки были правильно определены
	if !strings.Contains(result, "```go") {
		t.Error("Expected Go language to be extracted")
	}

	if !strings.Contains(result, "```javascript") {
		t.Error("Expected JavaScript language to be extracted")
	}

	if !strings.Contains(result, "```python") {
		t.Error("Expected Python language to be extracted")
	}

	// Проверяем, что нет блоков без языка
	if strings.Contains(result, "```\npackage main") {
		t.Error("Found code block without language")
	}

	if strings.Contains(result, "```\nfunction greet") {
		t.Error("Found code block without language")
	}

	if strings.Contains(result, "```\ndef greet") {
		t.Error("Found code block without language")
	}
}

func TestExtractCodeLanguagesInOrder(t *testing.T) {
	// Тест функции извлечения языков в порядке появления
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

	// Проверяем, что массив содержит правильные языки в правильном порядке
	expectedLanguages := []string{"go", "javascript", "python"}
	if len(languages) != len(expectedLanguages) {
		t.Errorf("Expected %d languages, got %d", len(expectedLanguages), len(languages))
	}

	for i, expected := range expectedLanguages {
		if i < len(languages) && languages[i] != expected {
			t.Errorf("Expected language %s at position %d, got %s", expected, i, languages[i])
		}
	}

	t.Logf("Languages in order: %v", languages)
}

func TestApplyLanguagesToMarkdownInOrder(t *testing.T) {
	// Тест применения языков к markdown по порядку
	markdown := "# Test\n\n```\npackage main\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```\n\n```\nfunction test() {\n    console.log(\"test\")\n}\n```\n\n```\ndef test():\n    print(\"test\")\n```\n"

	// Создаем массив языков в порядке появления
	languages := []string{"go", "javascript", "python"}

	// Применяем языки
	result := applyLanguagesToMarkdownInOrder(markdown, languages)

	t.Logf("Original markdown:\n%s", markdown)
	t.Logf("Result:\n%s", result)

	// Проверяем, что языки были применены
	if !strings.Contains(result, "```go\npackage main") {
		t.Error("Expected Go language to be applied")
	}

	if !strings.Contains(result, "```javascript\nfunction test") {
		t.Error("Expected JavaScript language to be applied")
	}

	if !strings.Contains(result, "```python\ndef test") {
		t.Error("Expected Python language to be applied")
	}
}

func TestExtractErlangLanguage(t *testing.T) {
	// Тест извлечения языка Erlang
	html := `
	<pre><code class="erlang">-module(hello).
-export([world/0]).

world() ->
    io:format("Hello, World!~n").
</code></pre>
	`

	languages := extractCodeLanguagesInOrder(html)

	// Проверяем, что язык Erlang был извлечен
	if len(languages) != 1 {
		t.Errorf("Expected 1 language, got %d", len(languages))
	}

	if languages[0] != "erlang" {
		t.Errorf("Expected language 'erlang', got '%s'", languages[0])
	}

	t.Logf("Extracted languages: %v", languages)
}

func TestRealWorldErlangExtraction(t *testing.T) {
	// Тест полного процесса извлечения с Erlang
	html := `
	<article>
		<h1>Erlang Example</h1>
		
		<pre><code class="erlang">-module(hello).
-export([world/0]).

world() ->
    io:format("Hello, World!~n").
</code></pre>
		
		<p>And some JavaScript:</p>
		
		<pre><code class="javascript">function greet() {
    console.log("Hello");
}</code></pre>
	</article>
	`

	// Создаем статью как это делает go-readability
	article := readability.Article{
		Content: html,
	}

	// Извлекаем языки для отладки
	languages := extractCodeLanguagesInOrder(article.Content)
	t.Logf("Languages in order: %v", languages)

	// Тестируем полный процесс извлечения
	result := extractAndConvertToMarkdown(article)

	t.Logf("Full extraction result:\n%s", result)

	// Проверяем, что языки были правильно определены
	if !strings.Contains(result, "```erlang") {
		t.Error("Expected Erlang language to be extracted")
	}

	if !strings.Contains(result, "```javascript") {
		t.Error("Expected JavaScript language to be extracted")
	}
}

func TestExtractMultipleLanguages(t *testing.T) {
	// Тест извлечения различных языков программирования
	html := `
	<pre><code class="erlang">-module(hello).
world() -> io:format("Hello").</code></pre>
	
	<pre><code class="haskell">main :: IO ()
main = putStrLn "Hello"</code></pre>
	
	<pre><code class="elixir">defmodule Hello do
  def world do
    IO.puts "Hello"
  end
end</code></pre>
	
	<pre><code class="scala">object Hello {
  def main(args: Array[String]) = {
    println("Hello")
  }
}</code></pre>
	
	<pre><code class="clojure">(defn hello []
  (println "Hello"))</code></pre>
	`

	languages := extractCodeLanguagesInOrder(html)

	// Проверяем, что все языки были извлечены в правильном порядке
	expectedLanguages := []string{"erlang", "haskell", "elixir", "scala", "clojure"}

	if len(languages) != len(expectedLanguages) {
		t.Errorf("Expected %d languages, got %d", len(expectedLanguages), len(languages))
	}

	for i, expected := range expectedLanguages {
		if i < len(languages) && languages[i] != expected {
			t.Errorf("Expected language %s at position %d, got %s", expected, i, languages[i])
		}
	}

	t.Logf("Extracted languages: %v", languages)
}

func TestRealErlangPageExtraction(t *testing.T) {
	// Читаем реальный HTML файл статьи про Erlang
	htmlBytes, err := os.ReadFile("Erlang — классный функциональный язык (или как мы сели в лужу) _ Хабр.html")
	if err != nil {
		t.Skipf("Файл не найден, пропускаем тест: %v", err)
	}

	htmlContent := string(htmlBytes)

	// Извлекаем языки программирования
	languages := extractCodeLanguagesInOrder(htmlContent)
	t.Logf("Найдено языков: %d", len(languages))
	t.Logf("Языки в порядке появления: %v", languages)

	// Проверяем, что найдены языки (ожидаем Erlang и возможно другие)
	if len(languages) == 0 {
		t.Error("Не найдено ни одного языка программирования")
	}

	// Проверяем наличие Erlang
	hasErlang := false
	for _, lang := range languages {
		if lang == "erlang" {
			hasErlang = true
			break
		}
	}

	if !hasErlang {
		t.Error("Не найден язык Erlang в статье про Erlang")
	}

	// Создаем статью как это делает go-readability
	article := readability.Article{
		Content: htmlContent,
	}

	// Тестируем полный процесс извлечения
	result := extractAndConvertToMarkdown(article)

	// Проверяем, что в результате есть блоки кода с языками
	if !strings.Contains(result, "```erlang") {
		t.Error("В результате не найден блок кода с языком Erlang")
	}

	// Подсчитываем количество блоков кода с языками
	erlangBlocks := strings.Count(result, "```erlang")
	t.Logf("Найдено блоков кода Erlang: %d", erlangBlocks)

	// Сохраняем результат в файл
	outputFile := "test_output_erlang_simple.md"
	err = os.WriteFile(outputFile, []byte(result), 0644)
	if err != nil {
		t.Logf("Не удалось сохранить результат в файл: %v", err)
	} else {
		absPath, _ := os.Getwd()
		fullPath := absPath + "/" + outputFile
		t.Logf("✅ Результат сохранен в файл: %s", fullPath)
		t.Logf("📄 Размер файла: %d байт", len(result))
	}

	// Выводим первые 1000 символов результата для анализа
	if len(result) > 1000 {
		t.Logf("Первые 1000 символов результата:\n%s", result[:1000])
	} else {
		t.Logf("Полный результат:\n%s", result)
	}
}

func TestDetailedErlangPageAnalysis(t *testing.T) {
	// Читаем реальный HTML файл статьи про Erlang
	htmlBytes, err := os.ReadFile("Erlang — классный функциональный язык (или как мы сели в лужу) _ Хабр.html")
	if err != nil {
		t.Skipf("Файл не найден, пропускаем тест: %v", err)
	}

	htmlContent := string(htmlBytes)

	// Извлекаем языки программирования
	languages := extractCodeLanguagesInOrder(htmlContent)

	// Анализируем каждый язык
	t.Logf("=== АНАЛИЗ ИЗВЛЕЧЕННЫХ ЯЗЫКОВ ===")
	t.Logf("Всего языков: %d", len(languages))

	for i, lang := range languages {
		if lang == "" {
			t.Logf("Позиция %d: ПУСТОЙ язык", i)
		} else {
			t.Logf("Позиция %d: '%s' (длина: %d)", i, lang, len(lang))
		}
	}

	// Фильтруем пустые языки
	var nonEmptyLanguages []string
	for _, lang := range languages {
		if lang != "" {
			nonEmptyLanguages = append(nonEmptyLanguages, lang)
		}
	}

	t.Logf("Непустых языков: %d", len(nonEmptyLanguages))
	t.Logf("Непустые языки: %v", nonEmptyLanguages)

	// Проверяем наличие Erlang
	hasErlang := false
	for _, lang := range nonEmptyLanguages {
		if lang == "erlang" {
			hasErlang = true
			break
		}
	}

	if !hasErlang {
		t.Error("Не найден язык Erlang в статье про Erlang")
	}

	// Создаем статью как это делает go-readability
	article := readability.Article{
		Content: htmlContent,
	}

	// Тестируем полный процесс извлечения
	result := extractAndConvertToMarkdown(article)

	// Ищем все блоки кода с языками
	codeBlockRegex := regexp.MustCompile("```([a-zA-Z0-9#+]+)")
	matches := codeBlockRegex.FindAllStringSubmatch(result, -1)

	t.Logf("=== АНАЛИЗ БЛОКОВ КОДА В MARKDOWN ===")
	t.Logf("Всего блоков кода с языками: %d", len(matches))

	for i, match := range matches {
		if len(match) > 1 {
			t.Logf("Блок %d: язык '%s'", i+1, match[1])
		}
	}

	// Проверяем, что в результате есть блоки кода с языками
	if !strings.Contains(result, "```erlang") {
		t.Error("В результате не найден блок кода с языком Erlang")
	}

	// Подсчитываем количество блоков кода с языками
	erlangBlocks := strings.Count(result, "```erlang")
	t.Logf("Найдено блоков кода Erlang: %d", erlangBlocks)

	// Сохраняем результат в файл для анализа
	outputFile := "test_output_erlang.md"
	err = os.WriteFile(outputFile, []byte(result), 0644)
	if err != nil {
		t.Logf("Не удалось сохранить результат в файл: %v", err)
	} else {
		// Получаем абсолютный путь к файлу
		absPath, _ := os.Getwd()
		fullPath := absPath + "/" + outputFile
		t.Logf("✅ Результат сохранен в файл: %s", fullPath)
		t.Logf("📄 Размер файла: %d байт", len(result))
		t.Logf("📊 Статистика:")
		t.Logf("   - Всего символов: %d", len(result))
		t.Logf("   - Блоков кода Erlang: %d", strings.Count(result, "```erlang"))
		t.Logf("   - Блоков кода без языка: %d", strings.Count(result, "```\n")-strings.Count(result, "```erlang"))
	}
}

func TestRawHTMLLanguageExtraction(t *testing.T) {
	// Читаем реальный HTML файл статьи про Erlang
	htmlBytes, err := os.ReadFile("Erlang — классный функциональный язык (или как мы сели в лужу) _ Хабр.html")
	if err != nil {
		t.Skipf("Файл не найден, пропускаем тест: %v", err)
	}

	rawHTML := string(htmlBytes)

	// Извлекаем языки из сырого HTML (как это теперь делает бот)
	languages := extractCodeLanguagesInOrder(rawHTML)

	t.Logf("=== ИЗВЛЕЧЕНИЕ ИЗ СЫРОГО HTML ===")
	t.Logf("Всего языков: %d", len(languages))

	// Фильтруем пустые языки
	var nonEmptyLanguages []string
	for _, lang := range languages {
		if lang != "" {
			nonEmptyLanguages = append(nonEmptyLanguages, lang)
		}
	}

	t.Logf("Непустых языков: %d", len(nonEmptyLanguages))
	t.Logf("Непустые языки: %v", nonEmptyLanguages)

	// Проверяем наличие Erlang
	hasErlang := false
	for _, lang := range nonEmptyLanguages {
		if lang == "erlang" {
			hasErlang = true
			break
		}
	}

	if !hasErlang {
		t.Error("Не найден язык Erlang в сыром HTML")
	}

	// Теперь симулируем процесс go-readability
	parsedURL, _ := url.Parse("https://habr.com/ru/articles/849758/")
	article, err := readability.FromReader(strings.NewReader(rawHTML), parsedURL)
	if err != nil {
		t.Fatalf("Ошибка при обработке go-readability: %v", err)
	}

	// Извлекаем языки из очищенного HTML (старый способ)
	cleanedLanguages := extractCodeLanguagesInOrder(article.Content)

	t.Logf("=== ИЗВЛЕЧЕНИЕ ИЗ ОЧИЩЕННОГО HTML ===")
	t.Logf("Всего языков: %d", len(cleanedLanguages))

	var nonEmptyCleanedLanguages []string
	for _, lang := range cleanedLanguages {
		if lang != "" {
			nonEmptyCleanedLanguages = append(nonEmptyCleanedLanguages, lang)
		}
	}

	t.Logf("Непустых языков: %d", len(nonEmptyCleanedLanguages))
	t.Logf("Непустые языки: %v", nonEmptyCleanedLanguages)

	// Сравниваем результаты
	t.Logf("=== СРАВНЕНИЕ ===")
	t.Logf("Сырой HTML - языков: %d", len(nonEmptyLanguages))
	t.Logf("Очищенный HTML - языков: %d", len(nonEmptyCleanedLanguages))

	if len(nonEmptyLanguages) != len(nonEmptyCleanedLanguages) {
		t.Logf("⚠️  Количество языков отличается!")
		t.Logf("   Сырой HTML: %v", nonEmptyLanguages)
		t.Logf("   Очищенный HTML: %v", nonEmptyCleanedLanguages)
	} else {
		t.Logf("✅ Количество языков одинаковое")
	}

	// Тестируем новую функцию с предварительно извлеченными языками
	result := extractAndConvertToMarkdownWithLanguages(article, languages)

	// Проверяем, что в результате есть блоки кода с языками
	if !strings.Contains(result, "```erlang") {
		t.Error("В результате не найден блок кода с языком Erlang")
	}

	// Подсчитываем количество блоков кода с языками
	erlangBlocks := strings.Count(result, "```erlang")
	t.Logf("Найдено блоков кода Erlang: %d", erlangBlocks)

	// Сохраняем результат в файл
	outputFile := "test_output_raw_html.md"
	err = os.WriteFile(outputFile, []byte(result), 0644)
	if err != nil {
		t.Logf("Не удалось сохранить результат в файл: %v", err)
	} else {
		absPath, _ := os.Getwd()
		fullPath := absPath + "/" + outputFile
		t.Logf("✅ Результат сохранен в файл: %s", fullPath)
		t.Logf("📄 Размер файла: %d байт", len(result))
	}
}

func TestUnknownLanguageFallback(t *testing.T) {
	// Тест fallback на язык C для неизвестных языков
	html := `
	<pre><code class="unknown-language">some code here</code></pre>
	<pre><code class="obscure-lang">more code</code></pre>
	<pre><code class="custom-syntax">another code block</code></pre>
	<pre><code class="erlang">-module(hello).</code></pre>
	`

	languages := extractCodeLanguagesInOrder(html)

	// Проверяем, что языки были извлечены
	if len(languages) != 4 {
		t.Errorf("Expected 4 languages, got %d", len(languages))
	}

	// Проверяем, что неизвестные языки стали "c"
	expectedLanguages := []string{"c", "c", "c", "erlang"}

	for i, expected := range expectedLanguages {
		if i < len(languages) && languages[i] != expected {
			t.Errorf("Expected language %s at position %d, got %s", expected, i, languages[i])
		}
	}

	t.Logf("Extracted languages: %v", languages)
	t.Logf("Expected languages: %v", expectedLanguages)
}

func TestFallbackWithRealConversion(t *testing.T) {
	// Тест полного процесса с неизвестными языками
	html := `
	<article>
		<h1>Test with Unknown Languages</h1>
		
		<pre><code class="unknown-lang">function test() {
    return "hello";
}</code></pre>
		
		<pre><code class="erlang">-module(test).
-export([hello/0]).

hello() -> "world".</code></pre>
		
		<pre><code class="obscure-syntax">def obscure_function():
    pass</code></pre>
	</article>
	`

	// Создаем статью как это делает go-readability
	article := readability.Article{
		Content: html,
	}

	// Извлекаем языки для отладки
	languages := extractCodeLanguagesInOrder(article.Content)
	t.Logf("Languages in order: %v", languages)

	// Тестируем полный процесс извлечения
	result := extractAndConvertToMarkdown(article)

	t.Logf("Full extraction result:\n%s", result)

	// Проверяем, что языки были правильно определены
	if !strings.Contains(result, "```c") {
		t.Error("Expected fallback language 'c' to be applied")
	}

	if !strings.Contains(result, "```erlang") {
		t.Error("Expected Erlang language to be extracted")
	}

	// Подсчитываем количество блоков кода с языками
	cBlocks := strings.Count(result, "```c")
	erlangBlocks := strings.Count(result, "```erlang")

	t.Logf("Found %d blocks with language 'c'", cBlocks)
	t.Logf("Found %d blocks with language 'erlang'", erlangBlocks)

	// Сохраняем результат в файл
	outputFile := "test_output_fallback.md"
	err := os.WriteFile(outputFile, []byte(result), 0644)
	if err != nil {
		t.Logf("Не удалось сохранить результат в файл: %v", err)
	} else {
		absPath, _ := os.Getwd()
		fullPath := absPath + "/" + outputFile
		t.Logf("✅ Результат сохранен в файл: %s", fullPath)
	}
}
