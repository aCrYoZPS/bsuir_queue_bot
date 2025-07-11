package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type MessagesService interface {
	HandleCommands(update *tgbotapi.Update, bot *tgbotapi.BotAPI)
}

type CallbacksService interface {
	HandleCallbacks(update *tgbotapi.Update, bot *tgbotapi.BotAPI)
}
