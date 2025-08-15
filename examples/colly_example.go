package main

import (
	"fmt"
	"strings"
	"time"

	"tgnip/internal"
)

func main() {
	// Пример URL для тестирования
	testURLs := []string{
		"https://habr.com/ru/articles/",
		"https://medium.com/",
		"https://dev.to/",
	}

	fmt.Println("=== Тестирование интеграции Colly с go-readability ===\n")

	for i, url := range testURLs {
		fmt.Printf("Тест %d: %s\n", i+1, url)
		fmt.Println("---")

		// Тест 1: Использование Colly по умолчанию
		fmt.Println("1. Использование Colly с настройками по умолчанию:")
		content, err := internal.ExtractContent(url)
		if err != nil {
			fmt.Printf("   Ошибка: %v\n", err)
		} else {
			fmt.Printf("   Заголовок: %s\n", content.Title)
			fmt.Printf("   Автор: %s\n", content.Author)
			fmt.Printf("   Дата: %s\n", content.Date)
			fmt.Printf("   Длина контента: %d символов\n", len(content.Markdown))
			if len(content.Markdown) > 100 {
				fmt.Printf("   Превью: %s...\n", content.Markdown[:100])
			}
		}

		fmt.Println()

		// Тест 2: Пользовательская конфигурация Colly
		fmt.Println("2. Использование Colly с пользовательской конфигурацией:")
		config := &internal.CollyConfig{
			UserAgent:      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			Timeout:        45 * time.Second,
			MaxRetries:     2,
			FollowRedirect: true,
			RespectRobots:  false,
		}
		content2, err := internal.ExtractContentWithConfig(url, config)
		if err != nil {
			fmt.Printf("   Ошибка: %v\n", err)
		} else {
			fmt.Printf("   Заголовок: %s\n", content2.Title)
			fmt.Printf("   Автор: %s\n", content2.Author)
			fmt.Printf("   Дата: %s\n", content2.Date)
			fmt.Printf("   Длина контента: %d символов\n", len(content2.Markdown))
		}

		fmt.Println()

		// Тест 3: Fallback механизм
		fmt.Println("3. Тест fallback механизма:")
		content3, err := internal.ExtractContentWithFallback(url)
		if err != nil {
			fmt.Printf("   Ошибка: %v\n", err)
		} else {
			fmt.Printf("   Заголовок: %s\n", content3.Title)
			fmt.Printf("   Автор: %s\n", content3.Author)
			fmt.Printf("   Дата: %s\n", content3.Date)
			fmt.Printf("   Длина контента: %d символов\n", len(content3.Markdown))
		}

		fmt.Println(strings.Repeat("=", 50))
		fmt.Println()
	}

	// Тест с невалидным URL
	fmt.Println("Тест с невалидным URL:")
	invalidURL := "https://invalid-domain-that-does-not-exist-12345.com/"
	content, err := internal.ExtractContentWithFallback(invalidURL)
	if err != nil {
		fmt.Printf("Ожидаемая ошибка: %v\n", err)
	} else {
		fmt.Printf("Неожиданный успех: %s\n", content.Title)
	}

	fmt.Println("\n=== Тестирование завершено ===")
}

// Пример функции для массового извлечения контента
func batchExtractContent(urls []string) {
	fmt.Println("Начинаю массовое извлечение контента...")

	results := make(chan *internal.Content, len(urls))
	errors := make(chan error, len(urls))

	// Запускаем горутины для параллельного извлечения
	for _, url := range urls {
		go func(u string) {
			content, err := internal.ExtractContentWithFallback(u)
			if err != nil {
				errors <- fmt.Errorf("ошибка для %s: %w", u, err)
				return
			}
			results <- content
		}(url)
	}

	// Собираем результаты
	successCount := 0
	errorCount := 0

	for i := 0; i < len(urls); i++ {
		select {
		case content := <-results:
			fmt.Printf("✓ Успешно извлечен контент: %s (%d символов)\n",
				content.Title, len(content.Markdown))
			successCount++
		case err := <-errors:
			fmt.Printf("✗ Ошибка: %v\n", err)
			errorCount++
		}
	}

	fmt.Printf("\nРезультаты: %d успешно, %d ошибок\n", successCount, errorCount)
}
