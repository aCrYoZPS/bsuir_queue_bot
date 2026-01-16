package tgutils

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/constants"
	datastructures "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/data_structures"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MuxHandler interface {
	Handle(ctx context.Context, message *tgbotapi.Message) error
	Revert(ctx context.Context, message *tgbotapi.Message) error
}

type HandlerFunc struct {
	handle func(ctx context.Context, message *tgbotapi.Message) error
	revert func(ctx context.Context, message *tgbotapi.Message) error
}

func NewHandlerFunc(handle, revert func(ctx context.Context, message *tgbotapi.Message) error) HandlerFunc {
	return HandlerFunc{handle: handle, revert: revert}
}
func (f HandlerFunc) Handle(ctx context.Context, message *tgbotapi.Message) error {
	return f.handle(ctx, message)
}
func (f HandlerFunc) Revert(ctx context.Context, message *tgbotapi.Message) error {
	return f.revert(ctx, message)
}

type CallbackHandler interface {
	HandleCallback(ctx context.Context, update *tgbotapi.Update, bot *Bot) error
}

type CallbackHandlerFunc func(ctx context.Context, update *tgbotapi.Update, bot *Bot) error

func (f CallbackHandlerFunc) HandleCallback(ctx context.Context, update *tgbotapi.Update, bot *Bot) error {
	return (func(ctx context.Context, update *tgbotapi.Update, bot *Bot) error)(f)(ctx, update, bot)
}

type Cache interface {
	SaveState(context.Context, interfaces.CachedInfo) error
	GetState(ctx context.Context, chatId int64) (*interfaces.CachedInfo, error)
	AcquireLock(ctx context.Context, chatId int64, key string) *sync.Mutex
	ReleaseLock(ctx context.Context, chatId int64, key string)
}

type Mux struct {
	routes          datastructures.TrieNode[MuxHandler]
	callbacks       datastructures.TrieNode[CallbackHandler]
	cache           Cache
	bot             *Bot
	NotFoundHandler MuxHandler
}

func NewMux(cache Cache, bot *Bot) *Mux {
	return &Mux{cache: cache, bot: bot, routes: datastructures.NewTrieNode[MuxHandler](),
		callbacks: datastructures.NewTrieNode[CallbackHandler](), NotFoundHandler: NewHandlerFunc(func(ctx context.Context, message *tgbotapi.Message) error { return errors.ErrUnsupported }, func(ctx context.Context, message *tgbotapi.Message) error { return errors.ErrUnsupported })}
}

func (mu *Mux) RegisterRoute(stateName string, handler MuxHandler) {
	mu.routes.Insert(stateName, handler)
}

func (mux *Mux) Handle(ctx context.Context, message *tgbotapi.Message) error {
	info, err := mux.cache.GetState(ctx, message.Chat.ID)
	if err != nil {
		return fmt.Errorf("couldn't get state in state machine: %w", err)
	}
	stateName := info.State()
	if message.Command() == strings.Trim(constants.REVERT_COMMAND, "/") {
		for route := range mux.routes.Iterate(stateName) {
			if route.IsLeaf() {
				err := route.Val().Revert(ctx, message)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	if _, ok := mux.routes.SearchExact(stateName); ok {
		for route := range mux.routes.Iterate(stateName) {
			if route.Val() != nil {
				err := route.Val().Handle(ctx, message)
				if err != nil {
					return err
				}
			}
		}
	} else {
		return mux.NotFoundHandler.Handle(ctx, message)
	}
	// if !routeFound {
	// 	return mux.routes.Search(constants.IDLE_STATE).Handle(ctx, message)
	// }
	return nil
}

func (mux *Mux) Revert(ctx context.Context, message *tgbotapi.Message) error {
	info, err := mux.cache.GetState(ctx, message.Chat.ID)
	if err != nil {
		return fmt.Errorf("couldn't get state in state machine: %w", err)
	}
	stateName := info.State()
	return mux.routes.Search(stateName).Revert(ctx, message)
}

func (mux *Mux) RegisterCallback(callbackName string, handler CallbackHandler) {
	mux.callbacks.Insert(callbackName, handler)
}

func (mux *Mux) HandleCallback(ctx context.Context, update *tgbotapi.Update, bot *Bot) error {
	return mux.callbacks.Search(update.CallbackData()).HandleCallback(ctx, update, bot)
}
