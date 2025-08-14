package main

import (
	"strings"
	"testing"
)

func TestLanguageDetectionIntegration(t *testing.T) {
	// Тестируем интеграцию детекции языка в markdown_converter
	content := &Content{
		Title:    "Test Article",
		Markdown: "# Content\n\nHere's some Go code:\n\n```\npackage main\n\nimport \"fmt\"\nimport \"time\"\n\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n    time.Sleep(time.Second)\n}\n```\n\nAnd some Python code:\n\n```\ndef hello():\n    print(\"Hello from Python!\")\n    return True\n```",
		Author:   "Test Author",
		Date:     "2024-01-01",
		URL:      "https://example.com/test",
	}

	locale := locales["en"]
	result := convertToMarkdown(content, "https://example.com/test", locale)

	// Проверяем, что языки были добавлены к блокам кода
	if !strings.Contains(result, "```go\n") {
		t.Error("Expected Go language to be detected and added")
	}

	if !strings.Contains(result, "```python\n") {
		t.Error("Expected Python language to be detected and added")
	}

	// Проверяем, что остальной контент сохранен
	if !strings.Contains(result, "Test Article") {
		t.Error("Expected title to be preserved")
	}

	if !strings.Contains(result, "Test Author") {
		t.Error("Expected author to be preserved")
	}

	t.Logf("Result contains Go: %v", strings.Contains(result, "```go\n"))
	t.Logf("Result contains Python: %v", strings.Contains(result, "```python\n"))
}
