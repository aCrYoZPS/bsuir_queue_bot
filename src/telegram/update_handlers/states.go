package update_handlers

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type State interface {
	StateName() string
	Handle(ctx context.Context, message *tgbotapi.Message) error
	Revert(ctx context.Context, message *tgbotapi.Message) error
}
