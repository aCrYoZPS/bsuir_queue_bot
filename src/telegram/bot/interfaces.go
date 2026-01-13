package bot

import (
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessagesService interface {
	HandleMessages(update *tgbotapi.Update, bot *tgutils.Bot)
}

type CallbacksService interface {
	HandleCallbacks(update *tgbotapi.Update, bot *tgutils.Bot)
}
