package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"botGastosPareja/internal/bot"
	"botGastosPareja/internal/config"
	"botGastosPareja/internal/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.NewDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize Telegram bot
	telegramBot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		log.Fatalf("Failed to initialize bot: %v", err)
	}

	log.Printf("Authorized on account %s", telegramBot.Self.UserName)

	// Create bot handler (commands are registered automatically)
	handler := bot.NewHandler(telegramBot, db)

	// Register commands with Telegram API
	if err := handler.RegisterTelegramCommands(); err != nil {
		log.Printf("Warning: Failed to register Telegram commands: %v", err)
	} else {
		log.Println("Bot commands registered with Telegram")
	}

	// Set up update configuration
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := telegramBot.GetUpdatesChan(u)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Println("Bot is running. Press Ctrl+C to stop.")

	// Process updates
	for {
		select {
		case update := <-updates:
			handler.HandleUpdate(update)
		case <-sigChan:
			log.Println("Shutting down...")
			return
		}
	}
}
