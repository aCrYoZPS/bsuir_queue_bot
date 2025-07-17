package stateMachine

import (
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AdminCallbackHandler struct {
	CallbackHandler
	repo interfaces.UsersRepository
}

func (handler *AdminCallbackHandler) HandleCallback(update *tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
	if _, err := bot.Request(callback); err != nil {
		return err
	}
	
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data)
	if _, err := bot.Send(msg); err != nil {
		return err
	}
	return nil
}
