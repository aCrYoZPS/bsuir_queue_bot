package bot

import (
	"fmt"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/logging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotController struct {
	bot         *tgbotapi.BotAPI
	msgSrv      MessagesService
	callbackSrv CallbacksService
}

func NewBotController(bot *tgbotapi.BotAPI, msgSrv MessagesService, callbackSrv CallbacksService) (*BotController, error) {
	bc := &BotController{
		bot:         bot,
		msgSrv:      msgSrv,
		callbackSrv: callbackSrv,
	}
	return bc, nil
}

func (controller *BotController) Start() {
	logging.Info(fmt.Sprintf("authorized on account %s", controller.bot.Self.UserName))
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	u.AllowedUpdates = []string{"message", "callback_query"}

	updates := controller.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if update.Message.Command() != "" {
				controller.msgSrv.HandleCommands(&update, controller.bot)
			} else if update.CallbackQuery != nil {
				controller.callbackSrv.HandleCallbacks(&update, controller.bot)
			} else {
				controller.msgSrv.HandleMessages(&update, controller.bot)
			}
		}
	}
}
