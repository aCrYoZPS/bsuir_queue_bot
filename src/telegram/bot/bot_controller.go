package bot

import (
	"errors"
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

func (bc *BotController) GetBot() (*tgbotapi.BotAPI, error) {
	if bc.bot == nil {
		return nil, errors.New("bot not initialized")
	}

	return bc.bot, nil
}

func (controller *BotController) Start() {
	bot, err := controller.GetBot()
	if err != nil {
		logging.FatalLog("failed to get bot instance", "error", err.Error())
	}

	logging.Info(fmt.Sprintf("authorized on account %s", bot.Self.UserName))

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if update.Message.Command() != "" {
				controller.msgSrv.HandleCommands(&update, bot)
			}
			if update.CallbackQuery != nil {
				controller.callbackSrv.HandleCallbacks(&update, bot)
			}
		}
	}
}
