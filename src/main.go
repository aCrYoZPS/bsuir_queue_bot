package main

import (
	"os"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/ioc"
	logging "github.com/aCrYoZPS/bsuir_queue_bot/src/logging"

	"github.com/joho/godotenv"
)

func main() {
	logging.InitLogging()

	_, exists := os.LookupEnv("ENVIRONMENT")
	if !exists {
		err := godotenv.Load()
		if err != nil {
			logging.FatalLog("Error loading .env file")
		}
	}
	controller := ioc.UseBotController()
	controller.Start()
}
