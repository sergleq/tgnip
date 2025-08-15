package internal

import (
	"testing"
)

func TestLanguageDetector(t *testing.T) {
	detector := NewLanguageDetector()

	tests := []struct {
		name     string
		markdown string
		expected []string
	}{
		{
			name: "Go code with explicit language",
			markdown: `# Go Example
			
Here's a Go example:

` + "```" + `go
package internal

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
` + "```" + ``,
			expected: []string{"go"},
		},
		{
			name: "Python code with context hints",
			markdown: `# Python Example
			
In Python, you can use the following code:

` + "```" + `
def hello_world():
    print("Hello, World!")
` + "```" + `

You can run this with: python script.py`,
			expected: []string{"python"},
		},
		{
			name: "JavaScript code with file extensions",
			markdown: `# JavaScript Example
			
Create a file called main.js:

` + "```" + `
const message = "Hello, World!";
console.log(message);
` + "```" + `

And run with: node main.js`,
			expected: []string{"javascript"},
		},
		{
			name: "Rust code with commands",
			markdown: `# Rust Example
			
` + "```" + `
fn main() {
    println!("Hello, World!");
}
` + "```" + `

Compile and run with: cargo run`,
			expected: []string{"rust"},
		},
		{
			name: "SQL code",
			markdown: `# Database Query
			
` + "```" + `
SELECT * FROM users WHERE active = true;
` + "```" + ``,
			expected: []string{"sql"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks, err := detector.DetectLanguagesFromMarkdown(tt.markdown)
			if err != nil {
				t.Fatalf("DetectLanguagesFromMarkdown() error = %v", err)
			}

			if len(blocks) == 0 {
				t.Error("No code blocks detected")
				return
			}

			for i, block := range blocks {
				score := detector.DetectLanguage(block)

				t.Logf("Block %d: Language=%s, Score=%.2f, Confidence=%.2f",
					i, score.Language, score.Score, score.Confidence)

				// Проверяем, что детектированный язык есть в ожидаемых
				found := false
				for _, expected := range tt.expected {
					if score.Language == expected {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("Expected one of %v, got %s", tt.expected, score.Language)
				}
			}
		})
	}
}

func TestContextSignals(t *testing.T) {
	detector := NewLanguageDetector()

	tests := []struct {
		name     string
		context  string
		expected string
	}{
		{
			name:     "Go context",
			context:  "In Go, you can use goroutines for concurrency",
			expected: "go",
		},
		{
			name:     "Python context",
			context:  "Using Python with Django framework",
			expected: "python",
		},
		{
			name:     "JavaScript context",
			context:  "In JavaScript, you can use async/await",
			expected: "javascript",
		},
		{
			name:     "Rust context",
			context:  "In Rust, you can use the actix framework",
			expected: "rust",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scores := make(map[string]float64)
			detector.addContextSignals(tt.context, scores)

			// Находим язык с максимальным скором
			var bestLanguage string
			var maxScore float64
			for lang, score := range scores {
				if score > maxScore {
					maxScore = score
					bestLanguage = lang
				}
			}

			if bestLanguage != tt.expected {
				t.Errorf("Expected %s, got %s (scores: %v)", tt.expected, bestLanguage, scores)
			}
		})
	}
}

func TestCodeHeuristics(t *testing.T) {
	detector := NewLanguageDetector()

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "Go code",
			content: `package main

import "fmt"

func main() {
    message := "Hello, World!"
    fmt.Println(message)
}`,
			expected: "go",
		},
		{
			name: "Python code",
			content: `def hello_world():
    print("Hello, World!")

if __name__ == "__main__":
    hello_world()`,
			expected: "python",
		},
		{
			name: "JavaScript code",
			content: `const message = "Hello, World!";
console.log(message);

async function fetchData() {
    const response = await fetch('/api/data');
    return response.json();
}`,
			expected: "javascript",
		},
		{
			name: "Rust code",
			content: `fn main() {
    let message = "Hello, World!";
    println!("{}", message);
}`,
			expected: "rust",
		},
		{
			name: "SQL code",
			content: `SELECT name, email 
FROM users 
WHERE active = true 
ORDER BY created_at DESC;`,
			expected: "sql",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scores := make(map[string]float64)
			detector.addCodeHeuristicSignals(tt.content, scores)

			// Находим язык с максимальным скором
			var bestLanguage string
			var maxScore float64
			for lang, score := range scores {
				if score > maxScore {
					maxScore = score
					bestLanguage = lang
				}
			}

			if bestLanguage != tt.expected {
				t.Errorf("Expected %s, got %s (scores: %v)", tt.expected, bestLanguage, scores)
			}
		})
	}
}
