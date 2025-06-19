package main

import (
	"log"
	"os"

	bot "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/bot"
	update_handlers "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	bot_token := os.Getenv("BOT_TOKEN")

	bot_controller := bot.GetBotController()

	err = bot_controller.Init(bot_token, true)
	if err != nil {
		log.Panic(err)
	}

	bot, err := bot_controller.GetBot()
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		update_handlers.HandleUpdate(&update, bot)
	}
}
