package ioc

import (
	"os"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/logging"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/bot"
)

var UseBotController = provider(
	func() *bot.BotController {
		bot_token := os.Getenv("BOT_TOKEN")
		bot, err := bot.NewBotController(bot_token, true, UseMessageService(), UseCallbacksService())
		if err != nil {
			logging.FatalLog(err.Error())
		}
		return bot
	},
)
