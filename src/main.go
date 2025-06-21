package main

import (
	"context"
	"fmt"

	gdocs_auth "github.com/aCrYoZPS/bsuir_queue_bot/src/google_docs/auth"
	iis_api "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api"
	logging "github.com/aCrYoZPS/bsuir_queue_bot/src/logging"
	bot "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/bot"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"github.com/joho/godotenv"
)

func main() {
	logging.InitLogging()
	err := godotenv.Load()
	if err != nil {
		logging.FatalLog("Error loading .env file")
	}

	// Perhaps, should handle API limits
	_, err = iis_api.GetAllGroups()
	if err != nil {
		logging.FatalLog(err.Error())
	}

	client, err := gdocs_auth.GetClient()
	if err != nil {
		logging.FatalLog(fmt.Sprintf("Unable to get google docs client: %v", err.Error()))
	}

	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		logging.FatalLog(fmt.Sprintf("Unable to retrieve sheets client: %v", err.Error()))
	}
	// Change later to do smth. As of now I am going to shop, so, whatever
	print(srv)
	bot.InitBot()
}
