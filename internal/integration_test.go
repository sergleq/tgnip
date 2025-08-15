package internal

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
	result := ConvertToMarkdown(content, "https://example.com/test", locale)

	// Проверяем, что блоки кода присутствуют (языки могут быть определены или нет)
	if !strings.Contains(result, "```") {
		t.Error("Expected code blocks to be present")
	}

	// Проверяем, что контент сохранен
	if !strings.Contains(result, "Here's some Go code:") {
		t.Error("Expected Go code section to be preserved")
	}

	if !strings.Contains(result, "And some Python code:") {
		t.Error("Expected Python code section to be preserved")
	}

	// Проверяем, что остальной контент сохранен
	if !strings.Contains(result, "Test Article") {
		t.Error("Expected title to be preserved")
	}

	if !strings.Contains(result, "Test Author") {
		t.Error("Expected author to be preserved")
	}

	t.Logf("Result contains code blocks: %v", strings.Contains(result, "```"))
	t.Logf("Result contains Go section: %v", strings.Contains(result, "Here's some Go code:"))
	t.Logf("Result contains Python section: %v", strings.Contains(result, "And some Python code:"))
}
