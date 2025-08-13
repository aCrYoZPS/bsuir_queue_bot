package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/ioc"
	logging "github.com/aCrYoZPS/bsuir_queue_bot/src/logging"

	"github.com/joho/godotenv"
)

func main() {
	logging.InitLogging()

	_, exists := os.LookupEnv("ENVIRONMENT")
	if !exists {
		err := godotenv.Load("../.env.local")
		if err != nil {
			logging.FatalLog("Error loading .env file")
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	controller := ioc.UseBotController()
	controller.Start(ctx)

	tasks := ioc.UseTasksController()
	tasks.InitTasks(ctx)

	ioc.Reset()
}
