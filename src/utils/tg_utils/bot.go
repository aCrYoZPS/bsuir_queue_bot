package tgutils

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	*tgbotapi.BotAPI
}

func NewBot(botApi *tgbotapi.BotAPI) *Bot {
	return &Bot{BotAPI: botApi}
}

func (bot *Bot) SendCtx(ctx context.Context, c tgbotapi.Chattable) (tgbotapi.Message, error) {
	resChan := make(chan struct {
		tgbotapi.Message
		error
	})
	go func(chan struct {
		tgbotapi.Message
		error
	}) {
		msg, err := bot.BotAPI.Send(c)
		resChan <- struct {
			tgbotapi.Message
			error
		}{msg, err}
	}(resChan)
	select {
	case res := <-resChan:
		return res.Message, res.error
	case <-ctx.Done():
		return tgbotapi.Message{}, ctx.Err()
	}
}
