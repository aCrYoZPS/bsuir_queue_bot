package update_handlers

import (
	"context"
	"log/slog"
	"runtime/debug"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/constants"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CallbacksService struct {
	cache   interfaces.HandlersCache
	handler CallbackHandler
}

func NewCallbackService(cache interfaces.HandlersCache, handler CallbackHandler) *CallbacksService {
	return &CallbacksService{
		cache:   cache,
		handler: handler,
	}
}

type CallbackHandler interface {
	HandleCallback(ctx context.Context, update *tgbotapi.Update, bot *tgutils.Bot) error
}

func (serv *CallbacksService) HandleCallbacks(update *tgbotapi.Update, bot *tgutils.Bot) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Recovered from panic", "err", r)
			debug.PrintStack()
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), constants.DEFAULT_TIMEOUT)
	defer cancel()
	if update.CallbackQuery == nil {
		slog.Error("no callback in update")
		return
	}
	msg := update.CallbackQuery.Message
	if msg == nil {
		return
	}
	mu := serv.cache.AcquireLock(ctx, msg.Chat.ID, update.CallbackData())
	locked := mu.TryLock()
	if !locked {
		return
	}

	defer mu.Unlock()
	defer serv.cache.ReleaseLock(ctx, msg.Chat.ID, update.CallbackData())

	err := serv.handler.HandleCallback(ctx, update, bot)
	if err != nil {
		slog.Error(err.Error())
	}
}
