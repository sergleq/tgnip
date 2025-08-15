package main

import (
	"fmt"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, using system environment variables")
	}

	// Get bot token
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		fmt.Println("Error: TELEGRAM_BOT_TOKEN not set")
		os.Exit(1)
	}

	// Create bot instance
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		fmt.Printf("Error creating bot: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Bot %s connected\n", bot.Self.UserName)

	// Delete webhook
	_, err = bot.Request(tgbotapi.DeleteWebhookConfig{})
	if err != nil {
		fmt.Printf("Error deleting webhook: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Webhook deleted successfully!")
	fmt.Println("You can now run the bot in polling mode.")
}
