package main

import (
	"github.com/aCrYoZPS/bsuir_queue_bot/src/ioc"
	logging "github.com/aCrYoZPS/bsuir_queue_bot/src/logging"
	bot "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/bot"

	"github.com/joho/godotenv"
)

func main() {
	logging.InitLogging()

	err := godotenv.Load()
	if err != nil {
		logging.FatalLog("Error loading .env file")
	}

	srv := ioc.UseSheetsApiService()
	err = srv.CreateSheets()
	if err != nil {
		logging.Error("failed to init sheets", "err", err.Error())
	}

	bot.InitBot()
}
