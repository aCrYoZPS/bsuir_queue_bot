package bot

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	update_handlers "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotController struct {
	bot *tgbotapi.BotAPI
}

func GetBotController() *BotController {
	return &BotController{
		bot: nil,
	}
}

func (bc *BotController) InitBotController(token string, debug bool) error {
	if bc.bot != nil {
		return errors.New("Invalid behaviour: tried to initialize bot twice")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return err
	}

	bc.bot = bot
	bc.bot.Debug = debug
	return nil
}

func (bc *BotController) GetBot() (*tgbotapi.BotAPI, error) {
	if bc.bot == nil {
		return nil, errors.New("Bot not initialized")
	}

	return bc.bot, nil
}

func InitBot() {
	bot_token := os.Getenv("BOT_TOKEN")

	bot_controller := GetBotController()

	err := bot_controller.InitBotController(bot_token, true)
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
