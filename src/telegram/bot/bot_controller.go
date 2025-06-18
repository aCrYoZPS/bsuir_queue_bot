package bot

import (
	"errors"

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

func (bc *BotController) Init(token string, debug bool) error {
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
