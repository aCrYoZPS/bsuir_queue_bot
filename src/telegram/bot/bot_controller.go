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

func NewBotController(token string, debug bool, msgSrv MessagesService, callbackSrv CallbacksService) (*BotController, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	bc := &BotController{
		bot:         bot,
		msgSrv:      msgSrv,
		callbackSrv: callbackSrv,
	}
	bc.bot.Debug = debug

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
				update_handlers.HandleCommands(&update, bot)
			} else {
				update_handlers.HandleMessage(&update, bot)
			}
		} else if update.CallbackQuery != nil {
			update_handlers.HandleCallbacks(&update, bot)
		}
	}
}
