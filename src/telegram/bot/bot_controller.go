package bot

import (
	"errors"
	"fmt"
	"os"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/logging"
	update_handlers "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotController struct {
	bot *tgbotapi.BotAPI
}

func GetBotController(token string, debug bool) (*BotController, error) {
	bc := &BotController{}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	bc.bot = bot
	bc.bot.Debug = debug

	return bc, nil
}

func (bc *BotController) GetBot() (*tgbotapi.BotAPI, error) {
	if bc.bot == nil {
		return nil, errors.New("Bot not initialized")
	}

	return bc.bot, nil
}

func InitBot() {
	bot_token := os.Getenv("BOT_TOKEN")

	bot_controller, err := GetBotController(bot_token, true)
	if err != nil {
		logging.FatalLog("Failed to initialize bot controller", "error", err.Error())
	}

	bot, err := bot_controller.GetBot()
	if err != nil {
		logging.FatalLog("Failed to get bot instance", "error", err.Error())
	}

	logging.Info(fmt.Sprintf("Authorized on account %s", bot.Self.UserName))

	update_handlers.InitCommands()
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if update.Message.Command() != "" {
				update_handlers.HandleCommands(&update, bot)
			}
			if update.CallbackQuery != nil {
				update_handlers.HandleCallbacks(&update, bot)
			}
		}
	}
}
