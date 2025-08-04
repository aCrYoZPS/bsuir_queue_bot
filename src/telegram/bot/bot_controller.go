package bot

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/logging"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotController struct {
	bot         *tgutils.Bot
	msgSrv      MessagesService
	callbackSrv CallbacksService
}

func NewBotController(bot *tgutils.Bot, msgSrv MessagesService, callbackSrv CallbacksService) (*BotController, error) {
	bc := &BotController{
		bot:         bot,
		msgSrv:      msgSrv,
		callbackSrv: callbackSrv,
	}
	return bc, nil
}

func (controller *BotController) Server() http.Client {
	return *http.DefaultClient
}

func (controller *BotController) Start(ctx context.Context) {
	logging.Info(fmt.Sprintf("authorized on account %s", controller.bot.Self.UserName))
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	u.AllowedUpdates = []string{"message", "callback_query"}

	updates := controller.bot.GetUpdatesChan(u)

	wg := sync.WaitGroup{}
	select {
	case <-ctx.Done():
		controller.bot.StopReceivingUpdates()
		wg.Wait()
	default:
		for update := range updates {
			if update.Message != nil {
				wg.Add(1)
				if update.Message.Command() != "" {
					go func(*sync.WaitGroup) {
						controller.msgSrv.HandleCommands(&update, controller.bot)
						wg.Done()
					}(&wg)
				} else {
					go func(*sync.WaitGroup) {
						controller.msgSrv.HandleMessages(&update, controller.bot)
						wg.Done()
					}(&wg)
				}
			} else if update.CallbackQuery != nil {
				go func(*sync.WaitGroup) {
					controller.callbackSrv.HandleCallbacks(&update, controller.bot)
					wg.Done()
				}(&wg)
			}
		}
	}
}
