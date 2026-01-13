package tgutils

import (
	"context"
	"errors"
	"unicode/utf8"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	*tgbotapi.BotAPI
}

func NewBot(botApi *tgbotapi.BotAPI) *Bot {
	return &Bot{BotAPI: botApi}
}

const (
	tgMsgMaxCharacters     = 4096
	tgCaptionMaxCharacters = 1024
)

var ErrMsgInvalidLen = errors.New("tg message overflows max capacity of characters")

func (bot *Bot) SendCtx(ctx context.Context, c tgbotapi.Chattable) (tgbotapi.Message, error) {
	resChan := make(chan struct {
		tgbotapi.Message
		error
	})
	go func(chan struct {
		tgbotapi.Message
		error
	}) {
		switch msg := c.(type) {
		case tgbotapi.MessageConfig:
			if utf8.RuneCountInString(msg.Text) > tgMsgMaxCharacters {
				resChan <- struct {
					tgbotapi.Message
					error
				}{tgbotapi.Message{}, ErrMsgInvalidLen}
			}
		case tgbotapi.DocumentConfig:
			if utf8.RuneCountInString(msg.Caption) > tgCaptionMaxCharacters {
				resChan <- struct {
					tgbotapi.Message
					error
				}{tgbotapi.Message{}, ErrMsgInvalidLen}
			}
		case tgbotapi.PhotoConfig:
			if utf8.RuneCountInString(msg.Caption) > tgCaptionMaxCharacters {
				resChan <- struct {
					tgbotapi.Message
					error
				}{tgbotapi.Message{}, ErrMsgInvalidLen}
			}
		}
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
