package internal

import (
	"regexp"
	"strings"

	"github.com/src-d/enry/v2"
)

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

// DetectLanguagesFromMarkdown анализирует Markdown и определяет языки для блоков кода
func (ld *LanguageDetector) DetectLanguagesFromMarkdown(markdown string) ([]CodeBlock, error) {
	var codeBlocks []CodeBlock

	// Разбиваем на строки для контекста
	lines := strings.Split(markdown, "\n")

	// Регулярное выражение для поиска блоков кода
	codeBlockRegex := regexp.MustCompile(`(?s)` + "```" + `(\w*)\s*\n(.*?)\n` + "```" + ``)
	matches := codeBlockRegex.FindAllStringSubmatch(markdown, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			language := match[1]
			content := match[2]

			// Получаем контекст вокруг блока
			context := ld.getContextAroundBlock(lines, 0) // Упрощенная версия

			block := CodeBlock{
				Content:     content,
				Language:    language,
				Context:     context,
				LineNumber:  0, // Упрощенная версия
				HasExplicit: language != "",
			}

			codeBlocks = append(codeBlocks, block)
		}
	}

	return codeBlocks, nil
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

	// 1. Сигналы из контекста
	ld.addContextSignals(block.Context, scores)

	// 2. Сигналы из имен файлов
	ld.addFileExtensionSignals(block.Context, scores)

	// 3. Сигналы из команд
	ld.addCommandSignals(block.Context, scores)

	// 4. Эвристики по коду
	ld.addCodeHeuristicSignals(block.Content, scores)

	// 5. Результат внешнего детектора (enry)
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

	// Вычисляем confidence (нормализуем до [0..1])
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

// getContextAroundBlock получает контекст вокруг блока кода
func (ld *LanguageDetector) getContextAroundBlock(lines []string, lineNum int) string {
	start := max(0, lineNum-ld.contextWindowSize-1)
	end := min(len(lines), lineNum+ld.contextWindowSize)

	var contextLines []string
	for i := start; i < end; i++ {
		contextLines = append(contextLines, lines[i])
	}

	return strings.Join(contextLines, "\n")
}

// addContextSignals добавляет сигналы из контекста
func (ld *LanguageDetector) addContextSignals(context string, scores map[string]float64) {
	// Явные упоминания языков в заголовках
	contextPatterns := map[string]*regexp.Regexp{
		"go":         regexp.MustCompile(`(?i)\b(?:in\s+go|using\s+go|go\s+example|golang)\b`),
		"python":     regexp.MustCompile(`(?i)\b(?:in\s+python|using\s+python|python\s+example)\b`),
		"javascript": regexp.MustCompile(`(?i)\b(?:in\s+js|using\s+js|javascript\s+example|node\.js|javascript)\b`),
		"typescript": regexp.MustCompile(`(?i)\b(?:in\s+ts|using\s+ts|typescript\s+example)\b`),
		"rust":       regexp.MustCompile(`(?i)\b(?:in\s+rust|using\s+rust|rust\s+example)\b`),
		"java":       regexp.MustCompile(`(?i)\b(?:in\s+java|using\s+java|java\s+example)\b`),
		"c++":        regexp.MustCompile(`(?i)\b(?:in\s+c\+\+|using\s+c\+\+|cpp\s+example)\b`),
		"c":          regexp.MustCompile(`(?i)\b(?:in\s+c\b|using\s+c\b|c\s+example)\b`),
		"bash":       regexp.MustCompile(`(?i)\b(?:in\s+bash|using\s+bash|shell\s+script)\b`),
		"sql":        regexp.MustCompile(`(?i)\b(?:in\s+sql|using\s+sql|database\s+query)\b`),
	}

	for lang, pattern := range contextPatterns {
		if pattern.MatchString(context) {
			scores[lang] += ld.contextWeight
		}
	}

	// Фреймворки и библиотеки
	frameworkPatterns := map[string]*regexp.Regexp{
		"python":     regexp.MustCompile(`(?i)\b(?:django|flask|fastapi|pandas|numpy|requests)\b`),
		"java":       regexp.MustCompile(`(?i)\b(?:spring|maven|gradle|junit|hibernate)\b`),
		"go":         regexp.MustCompile(`(?i)\b(?:gin|fiber|echo|gorilla|cobra|viper)\b`),
		"javascript": regexp.MustCompile(`(?i)\b(?:express|nest|react|vue|angular|webpack)\b`),
		"typescript": regexp.MustCompile(`(?i)\b(?:express|nest|react|vue|angular|webpack)\b`),
		"rust":       regexp.MustCompile(`(?i)\b(?:actix|rocket|tokio|serde|clap)\b`),
	}

	for lang, pattern := range frameworkPatterns {
		if pattern.MatchString(context) {
			scores[lang] += ld.contextWeight * 0.5
		}
	}
}

// addFileExtensionSignals добавляет сигналы из расширений файлов
func (ld *LanguageDetector) addFileExtensionSignals(context string, scores map[string]float64) {
	filePatterns := map[string]*regexp.Regexp{
		"go":         regexp.MustCompile(`\b\w+\.go\b`),
		"python":     regexp.MustCompile(`\b\w+\.py\b`),
		"javascript": regexp.MustCompile(`\b\w+\.js\b`),
		"typescript": regexp.MustCompile(`\b\w+\.ts\b`),
		"rust":       regexp.MustCompile(`\b\w+\.rs\b`),
		"java":       regexp.MustCompile(`\b\w+\.java\b`),
		"c++":        regexp.MustCompile(`\b\w+\.(?:cpp|cc|cxx|hpp|h)\b`),
		"c":          regexp.MustCompile(`\b\w+\.(?:c|h)\b`),
		"bash":       regexp.MustCompile(`\b\w+\.(?:sh|bash)\b`),
		"sql":        regexp.MustCompile(`\b\w+\.sql\b`),
	}

	// Конфигурационные файлы
	configPatterns := map[string]*regexp.Regexp{
		"go":         regexp.MustCompile(`\b(?:go\.mod|go\.sum)\b`),
		"python":     regexp.MustCompile(`\b(?:requirements\.txt|setup\.py|pyproject\.toml)\b`),
		"javascript": regexp.MustCompile(`\b(?:package\.json|package-lock\.json|yarn\.lock)\b`),
		"typescript": regexp.MustCompile(`\b(?:package\.json|package-lock\.json|yarn\.lock|tsconfig\.json)\b`),
		"rust":       regexp.MustCompile(`\b(?:Cargo\.toml|Cargo\.lock)\b`),
		"java":       regexp.MustCompile(`\b(?:pom\.xml|build\.gradle)\b`),
	}

	for lang, pattern := range filePatterns {
		if pattern.MatchString(context) {
			scores[lang] += ld.fileExtWeight
		}
	}

	for lang, pattern := range configPatterns {
		if pattern.MatchString(context) {
			scores[lang] += ld.fileExtWeight * 0.8
		}
	}
}

// addCommandSignals добавляет сигналы из команд
func (ld *LanguageDetector) addCommandSignals(context string, scores map[string]float64) {
	commandPatterns := map[string]*regexp.Regexp{
		"go":         regexp.MustCompile(`\b(?:go\s+(?:build|run|mod|get|install)|go\b)`),
		"python":     regexp.MustCompile(`\b(?:python\s+(?:-m\s+venv|manage\.py|pip\s+install)|pip\s+install)\b`),
		"javascript": regexp.MustCompile(`\b(?:node\s+\w+\.js|npm\s+(?:i|install|run)|yarn\s+(?:add|install))\b`),
		"typescript": regexp.MustCompile(`\b(?:node\s+\w+\.js|npm\s+(?:i|install|run)|yarn\s+(?:add|install)|tsc\b)\b`),
		"rust":       regexp.MustCompile(`\b(?:cargo\s+(?:run|build|test|check)|rustc\b)\b`),
		"java":       regexp.MustCompile(`\b(?:javac\b|java\s+-jar|mvn\s+(?:compile|run)|gradle\s+(?:build|run))\b`),
		"c++":        regexp.MustCompile(`\b(?:g\+\+\b|clang\+\+\b|make\b|cmake\b)\b`),
		"c":          regexp.MustCompile(`\b(?:gcc\b|clang\b|make\b)\b`),
		"bash":       regexp.MustCompile(`\b(?:bash\b|sh\b|\.\/\w+\.sh)\b`),
	}

	for lang, pattern := range commandPatterns {
		if pattern.MatchString(context) {
			scores[lang] += ld.commandWeight
		}
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

	// Rust сигнатуры
	if strings.Contains(content, "fn main()") ||
		strings.Contains(content, "let mut") ||
		strings.Contains(content, "trait ") ||
		strings.Contains(content, "impl ") ||
		strings.Contains(content, "println!") ||
		(strings.Contains(content, "use ") && !strings.Contains(content, "import ")) {
		scores["rust"] += ld.heuristicWeight
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

	// TypeScript сигналы
	if strings.Contains(content, ": string") ||
		strings.Contains(content, ": number") ||
		strings.Contains(content, ": boolean") ||
		strings.Contains(content, "interface ") ||
		strings.Contains(content, "type ") {
		scores["typescript"] += ld.heuristicWeight
	}

	// C/C++ сигнатуры
	if strings.Contains(content, "#include <") ||
		strings.Contains(content, "int main()") ||
		strings.Contains(content, "std::") ||
		strings.Contains(content, "using namespace") {
		scores["c++"] += ld.heuristicWeight
	}

	// Java сигнатуры
	if strings.Contains(content, "public class") ||
		strings.Contains(content, "@Override") ||
		strings.Contains(content, "System.out.println") ||
		strings.Contains(content, "public static void main") {
		scores["java"] += ld.heuristicWeight
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

	// SQL сигнатуры
	if strings.Contains(content, "SELECT ") ||
		strings.Contains(content, "FROM ") ||
		strings.Contains(content, "WHERE ") ||
		strings.Contains(content, "JOIN ") ||
		strings.Contains(content, "INSERT INTO") {
		scores["sql"] += ld.heuristicWeight
	}

	// YAML/TOML/JSON сигнатуры
	if strings.Contains(content, "key: value") ||
		strings.Contains(content, "key = value") ||
		strings.Contains(content, "\"key\": \"value\"") {
		if strings.Contains(content, "---") || strings.Contains(content, ": ") {
			scores["yaml"] += ld.heuristicWeight
		} else if strings.Contains(content, " = ") {
			scores["toml"] += ld.heuristicWeight
		} else if strings.Contains(content, "\"") && strings.Contains(content, "{") {
			scores["json"] += ld.heuristicWeight
		}
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
				weight *= 0.5 // Меньший вес для коротких сниппетов
			} else if len(content) > 200 {
				weight *= 1.5 // Больший вес для длинных сниппетов
			}
			scores[mappedLang] += weight
		}
	}
}

// max возвращает максимальное из двух чисел
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min возвращает минимальное из двух чисел
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
