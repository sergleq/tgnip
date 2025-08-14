# Система детекции языка программирования

Система для автоматического определения языка программирования в блоках кода Markdown файлов.

## Возможности

- **Многоуровневая детекция**: использует несколько источников сигналов для точного определения языка
- **Контекстный анализ**: учитывает заголовки, подписи, команды и файлы вокруг блока кода
- **Эвристический анализ**: распознает характерные синтаксические конструкции языков
- **Внешний детектор**: интегрирует библиотеку enry для дополнительной точности
- **Взвешенная система**: присваивает разные веса различным типам сигналов

## Поддерживаемые языки

- **Go** - package main, func, import, :=, fmt.
- **Python** - def, import, self, if __name__, отступы
- **JavaScript** - import/export, async/await, const/let
- **TypeScript** - типы :string, interface, type
- **Rust** - fn main(), let mut, trait, impl, println!
- **Java** - public class, @Override, System.out.println
- **C/C++** - #include, int main(), std::
- **Bash** - shebang, команды cd, grep, sed
- **SQL** - SELECT, FROM, WHERE, JOIN
- **YAML/TOML/JSON** - структуры ключ-значение

## Архитектура

### Основные компоненты

1. **LanguageDetector** - основной класс детектора
2. **CodeBlock** - структура для представления блока кода с контекстом
3. **LanguageScore** - результат детекции с оценкой уверенности

### Система весов

- **Явное указание языка** (```go): вес 10.0
- **Контекстные сигналы** (заголовки, фреймворки): вес 3.0
- **Расширения файлов** (.go, .py, main.js): вес 3.0
- **Команды** (go build, python script.py): вес 2.0
- **Эвристики кода** (синтаксис): вес 2.0
- **Enry детектор** (внешний): вес 2.0

## Использование

### Базовое использование

```go
// Создаем детектор
detector := NewLanguageDetector()

// Анализируем Markdown
blocks, err := detector.DetectLanguagesFromMarkdown(markdown)
if err != nil {
    log.Fatal(err)
}

// Определяем язык для каждого блока
for _, block := range blocks {
    score := detector.DetectLanguage(block)
    fmt.Printf("Язык: %s, Уверенность: %.2f\n", 
        score.Language, score.Confidence)
}
```

### Пример с контекстом

```go
markdown := `# Go пример

В Go вы можете использовать горутины:

` + "```" + `go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `

Запустите с: go run main.go`

blocks, _ := detector.DetectLanguagesFromMarkdown(markdown)
score := detector.DetectLanguage(blocks[0])
// Результат: Language=go, Score=10.0, Confidence=1.0
```

## Алгоритм детекции

### 1. Извлечение блоков кода

Система использует регулярные выражения для поиска блоков кода в формате:
```
```language
code content
```
```

### 2. Анализ контекста

Для каждого блока анализируется:
- Заголовки и подзаголовки (±3 строки)
- Упоминания языков ("In Go...", "Using Python")
- Фреймворки и библиотеки (Django, Express, Gin)
- Имена файлов и расширения (main.go, requirements.txt)
- Команды установки/запуска (go build, npm install)

### 3. Эвристический анализ

Анализирует синтаксические особенности:
- **Go**: `package main`, `func`, `import "fmt"`, `:=`
- **Python**: `def`, `import`, `self`, отсутствие `;`
- **JavaScript**: `import`, `export`, `async/await`
- **Rust**: `fn main()`, `let mut`, `trait`, `impl`

### 4. Внешний детектор

Использует библиотеку enry для дополнительной проверки:
- Анализирует структуру и паттерны кода
- Учитывает длину сниппета (больший вес для длинных блоков)
- Маппинг результатов к поддерживаемым языкам

### 5. Вычисление результата

- Суммирует все сигналы с учетом весов
- Выбирает язык с максимальным скором
- Вычисляет confidence как отношение максимального скора к общей сумме

## Примеры детекции

### Явное указание языка
```markdown
```go
package main
```
```
Результат: go (Score: 10.0, Confidence: 1.0)

### Контекстные подсказки
```markdown
# Python пример
В Python с Django:

```
def view(request):
    return JsonResponse({"message": "Hello"})
```

Запустите: python manage.py runserver
```
Результат: python (Score: 5.0, Confidence: 1.0)

### Эвристический анализ
```markdown
```
fn main() {
    let message = "Hello, World!";
    println!("{}", message);
}
```
```
Результат: rust (Score: 2.0, Confidence: 1.0)

### Комбинированные сигналы
```markdown
Создайте файл main.js:

```
const express = require('express');
const app = express();

app.get('/', (req, res) => {
    res.json({ message: 'Hello!' });
});
```

Запустите: node main.js
```
Результат: javascript (Score: 8.0, Confidence: 1.0)

## Запуск примеров

```bash
# Запуск основного примера
go run cmd/language_detection_example/main.go

# Запуск тестов
go test -v -run TestLanguageDetector
go test -v -run TestContextSignals
go test -v -run TestCodeHeuristics
```

## Интеграция в проект

Система легко интегрируется в существующие проекты:

```go
// В content_extractor.go или markdown_converter.go
func processMarkdownContent(content string) {
    detector := NewLanguageDetector()
    blocks, _ := detector.DetectLanguagesFromMarkdown(content)
    
    for _, block := range blocks {
        score := detector.DetectLanguage(block)
        // Обработка блока с известным языком
        processCodeBlock(block.Content, score.Language)
    }
}
```

## Настройка весов

Можно настроить веса для различных типов сигналов:

```go
detector := &LanguageDetector{
    contextWeight:     3.0,  // Контекстные сигналы
    fileExtWeight:     3.0,  // Расширения файлов
    commandWeight:     2.0,  // Команды
    heuristicWeight:   2.0,  // Эвристики кода
    enryWeight:        2.0,  // Внешний детектор
    explicitWeight:    10.0, // Явное указание
    contextWindowSize: 3,    // Размер окна контекста
}
```

## Производительность

- **Парсинг**: O(n) где n - размер Markdown
- **Детекция языка**: O(m) где m - количество блоков кода
- **Память**: O(k) где k - размер самого большого блока кода

Система оптимизирована для быстрой работы с большими документами и может обрабатывать сотни блоков кода за секунду.
