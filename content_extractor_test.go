package main

import (
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

func TestProcessCodeBlocksInMarkdown(t *testing.T) {
	// Простой HTML с блоком кода
	html := `<pre><code class="language-go">package main

func main() {
    fmt.Println("Hello")
}</code></pre>`

	// Markdown без языка (как возвращает html-to-markdown)
	markdown := "```\npackage main\n\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```"

	// Ожидаемый результат
	expected := "```go\npackage main\n\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```"

	result := processCodeBlocksInMarkdown(markdown, html)

	if result != expected {
		t.Errorf("processCodeBlocksInMarkdown() = \n%q\nwant \n%q", result, expected)
	}
}

func TestProcessCodeBlocksInMarkdownComplex(t *testing.T) {
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

	// Markdown без языков (как возвращает html-to-markdown)
	markdown := "Some text before.\n\n" +
		"```\n" +
		"def hello():\n" +
		"    print(\"Hello from Python\")\n" +
		"```\n\n" +
		"Some text between.\n\n" +
		"```\n" +
		"function hello() {\n" +
		"    console.log(\"Hello from JS\");\n" +
		"}\n" +
		"```\n\n" +
		"Some text after."

	// Ожидаемый результат
	expected := "Some text before.\n\n" +
		"```python\n" +
		"def hello():\n" +
		"    print(\"Hello from Python\")\n" +
		"```\n\n" +
		"Some text between.\n\n" +
		"```javascript\n" +
		"function hello() {\n" +
		"    console.log(\"Hello from JS\");\n" +
		"}\n" +
		"```\n\n" +
		"Some text after."

	result := processCodeBlocksInMarkdown(markdown, html)

	if result != expected {
		t.Errorf("processCodeBlocksInMarkdown() = \n%q\nwant \n%q", result, expected)
		t.Logf("Actual result:\n%s", result)
		t.Logf("Expected result:\n%s", expected)
	}
}

func TestProcessCodeBlocksInMarkdownRealistic(t *testing.T) {
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

	// Markdown, который может вернуть html-to-markdown
	markdown := "# Programming Examples\n\n" +
		"Here are some code examples:\n\n" +
		"```\n" +
		"package main\n\n" +
		"import \"fmt\"\n\n" +
		"func main() {\n" +
		"    fmt.Println(\"Hello, World!\")\n" +
		"}\n" +
		"```\n\n" +
		"And here's some JavaScript:\n\n" +
		"```\n" +
		"function greet(name) {\n" +
		"    console.log(\"Hello, \" + name + \"!\");\n" +
		"}\n\n" +
		"greet(\"World\");\n" +
		"```\n\n" +
		"And Python:\n\n" +
		"```\n" +
		"def greet(name):\n" +
		"    print(f\"Hello, {name}!\")\n\n" +
		"greet(\"World\")\n" +
		"```"

	// Ожидаемый результат
	expected := "# Programming Examples\n\n" +
		"Here are some code examples:\n\n" +
		"```go\n" +
		"package main\n\n" +
		"import \"fmt\"\n\n" +
		"func main() {\n" +
		"    fmt.Println(\"Hello, World!\")\n" +
		"}\n" +
		"```\n\n" +
		"And here's some JavaScript:\n\n" +
		"```javascript\n" +
		"function greet(name) {\n" +
		"    console.log(\"Hello, \" + name + \"!\");\n" +
		"}\n\n" +
		"greet(\"World\");\n" +
		"```\n\n" +
		"And Python:\n\n" +
		"```python\n" +
		"def greet(name):\n" +
		"    print(f\"Hello, {name}!\")\n\n" +
		"greet(\"World\")\n" +
		"```"

	result := processCodeBlocksInMarkdown(markdown, html)

	if result != expected {
		t.Errorf("processCodeBlocksInMarkdown() = \n%q\nwant \n%q", result, expected)
		t.Logf("Actual result:\n%s", result)
		t.Logf("Expected result:\n%s", expected)
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

	// Обрабатываем блоки кода
	result := processCodeBlocksInMarkdown(markdown, html)

	t.Logf("After processing code blocks:\n%s", result)

	// Проверяем, что языки были добавлены
	if !strings.Contains(result, "```go") {
		t.Error("Expected Go language to be extracted")
	}

	if !strings.Contains(result, "```javascript") {
		t.Error("Expected JavaScript language to be extracted")
	}

	if !strings.Contains(result, "```python") {
		t.Error("Expected Python language to be extracted")
	}
}

func TestFixIncorrectLanguageBlocks(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		expected string
	}{
		{
			name:     "Fix highlight javascript",
			markdown: "```highlight javascript\nconsole.log('Hello');\n```",
			expected: "```javascript\nconsole.log('Hello');\n```",
		},
		{
			name:     "Fix code python",
			markdown: "```code python\ndef hello():\n    print('Hello')\n```",
			expected: "```python\ndef hello():\n    print('Hello')\n```",
		},
		{
			name:     "Fix source go",
			markdown: "```source go\npackage main\nfunc main() {}\n```",
			expected: "```go\npackage main\nfunc main() {}\n```",
		},
		{
			name:     "Fix hljs javascript",
			markdown: "```hljs javascript\nfunction test() {}\n```",
			expected: "```javascript\nfunction test() {}\n```",
		},
		{
			name:     "Keep correct language",
			markdown: "```javascript\nconsole.log('Hello');\n```",
			expected: "```javascript\nconsole.log('Hello');\n```",
		},
		{
			name:     "Keep block without language",
			markdown: "```\njust text\n```",
			expected: "```\njust text\n```",
		},
		{
			name:     "Fix multiple blocks",
			markdown: "```highlight javascript\nconsole.log('Hello');\n```\n\n```code python\nprint('Hello')\n```",
			expected: "```javascript\nconsole.log('Hello');\n```\n\n```python\nprint('Hello')\n```",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем пустой HTML, так как мы тестируем только исправление markdown
			html := "<article></article>"

			result := processCodeBlocksInMarkdown(tt.markdown, html)

			if result != tt.expected {
				t.Errorf("processCodeBlocksInMarkdown() = \n%q\nwant \n%q", result, tt.expected)
			}
		})
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

	// Проверяем, что языки были правильно определены
	if !strings.Contains(result, "```javascript") {
		t.Error("Expected JavaScript language to be extracted")
	}

	if !strings.Contains(result, "```python") {
		t.Error("Expected Python language to be extracted")
	}

	if !strings.Contains(result, "```go") {
		t.Error("Expected Go language to be extracted")
	}

	// Проверяем, что нет неправильных языков
	if strings.Contains(result, "```highlight") {
		t.Error("Found incorrect language 'highlight' in result")
	}

	if strings.Contains(result, "```language-") {
		t.Error("Found incorrect language prefix in result")
	}
}
