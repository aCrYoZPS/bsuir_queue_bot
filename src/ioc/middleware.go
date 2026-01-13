package ioc

import (
	"context"
	"fmt"
	"slices"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var useAdminMiddleware = func(next tgutils.MuxHandler) func() tgutils.MuxHandler {
	bot := useTgBot()
	users := useUsersRepository()
	return provider(
		func() tgutils.MuxHandler {
			return tgutils.NewHandlerFunc(
				func(ctx context.Context, message *tgbotapi.Message) error {
					user, err := users.GetByTgId(ctx, message.From.ID)
					if err != nil {
						return fmt.Errorf("failed to get user by id during admin middleware: %w", err)
					}
					if !slices.Contains(user.Roles, entities.Admin) {
						_, err := bot.SendCtx(ctx, tgbotapi.NewMessage(message.From.ID, "Вы не являетесь админом для выполнения этой команды"))
						if err != nil {
							return fmt.Errorf("failed to send not admin message during admin middleware handling: %w", err)
						}

					}
					return next.Handle(ctx, message)
				}, next.Revert)
		},
	)
}

// var usePanicMiddleware = func(next tgutils.MuxHandler) func() tgutils.MuxHandler {
// 	return provider(
// 		func() tgutils.MuxHandler {
// 			return tgutils.NewHandlerFunc(
// 				func(ctx context.Context, message *tgbotapi.Message) error {
// 					defer func() {
// 						if r := recover(); r != nil {
// 							slog.Error("Recovered from panic", "err", r)
// 							debug.PrintStack()
// 						}
// 					}()
// 					return next.Handle(ctx, message)
// 				},
// 				func(ctx context.Context, message *tgbotapi.Message) error {
// 					defer func() {
// 						if r := recover(); r != nil {
// 							slog.Error("Recovered from panic", "err", r)
// 							debug.PrintStack()
// 						}
// 					}()
// 					return next.Revert(ctx, message)
// 				},
// 			)
// 		},
// 	)
// }

// var useMutexMiddleware = func(next tgutils.MuxHandler) tgutils.MuxHandler {
// 	cache := useHandlersCache()
// 	return provider(
// 		func() tgutils.MuxHandler {
// 			return tgutils.NewHandlerFunc(
// 				func(ctx context.Context, message *tgbotapi.Message) error {
// 					mu := cache.AcquireLock(ctx, message.Chat.ID, "")
// 					mu.Lock()
// 					err := next.Handle(ctx, message)
// 					mu.Unlock()
// 					return err
// 				},
// 				func(ctx context.Context, message *tgbotapi.Message) error {
// 					mu := cache.AcquireLock(ctx, message.Chat.ID, "")
// 					mu.Lock()
// 					err := next.Revert(ctx, message)
// 					mu.Unlock()
// 					return err
// 				},
// 			)
// 		},
// 	)()
// }
