package stateMachine

import (
	"context"
	"log/slog"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/admin"
	adminInterfaces "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/admin/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/group"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CallbacksService struct {
	//More of a placeholder, which will contain inject google services to handle callbacks
	usersRepo     interfaces.UsersRepository
	requests      interfaces.RequestsRepository
	cache         interfaces.HandlersCache
	adminRequests interfaces.AdminRequestsRepository
	lessons       adminInterfaces.LessonsService
}

func NewCallbackService(usersRepo interfaces.UsersRepository, cache interfaces.HandlersCache,
	requests interfaces.RequestsRepository, adminRequests interfaces.AdminRequestsRepository,
	lessons adminInterfaces.LessonsService) *CallbacksService {
	return &CallbacksService{
		usersRepo:     usersRepo,
		cache:         cache,
		requests:      requests,
		adminRequests: adminRequests,
		lessons:       lessons,
	}
}

type CallbackHandler interface {
	HandleCallback(ctx context.Context, update *tgbotapi.Update, bot *tgbotapi.BotAPI) error
}

func (serv *CallbacksService) HandleCallbacks(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
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
	mu := serv.cache.AcquireLock(ctx, msg.Chat.ID)
	mu.Lock()

	defer mu.Unlock()
	defer serv.cache.ReleaseLock(ctx, msg.Chat.ID)

	var callback_handler CallbackHandler
	switch {
	case strings.HasPrefix(update.CallbackData(), constants.ADMIN_CALLBACKS):
		callback_handler = admin.NewAdminCallbackHandler(serv.usersRepo, serv.cache, serv.adminRequests, serv.lessons)
	case strings.HasPrefix(update.CallbackData(), constants.GROUP_CALLBACKS):
		callback_handler = group.NewGroupCallbackHandler(serv.usersRepo, serv.cache, serv.requests)
	}

	err := callback_handler.HandleCallback(ctx, update, bot)
	if err != nil {
		slog.Error(err.Error())
	}
}
