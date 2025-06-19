package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	iis_api "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api"
	bot "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/bot"
	update_handlers "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/joho/godotenv"
)

var opts = &slog.HandlerOptions{
	AddSource: false,
	Level:     slog.LevelDebug,
	ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
		if attr.Key != slog.TimeKey {
			return attr
		}

		curTime := attr.Value.Time()

		attr.Value = slog.StringValue(curTime.Format(time.DateTime))
		return attr
	},
}
var logger = slog.New(slog.NewTextHandler(os.Stderr, opts))

func main() {
	slog.SetDefault(logger)
	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file")
		os.Exit(-1)
	}

	_, err = iis_api.GetAllGroups()
	if err != nil {
		slog.Error(err.Error())
	}

	bot_token := os.Getenv("BOT_TOKEN")

	bot_controller := bot.GetBotController()

	err = bot_controller.Init(bot_token, true)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(-1)
	}

	bot, err := bot_controller.GetBot()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(-1)
	}

	slog.Info(fmt.Sprintf("Authorized on account %s", bot.Self.UserName))

	update_handlers.InitCommands()
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		update_handlers.HandleUpdate(&update, bot)
	}
}
