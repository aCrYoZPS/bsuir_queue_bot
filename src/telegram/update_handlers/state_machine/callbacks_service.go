package stateMachine

import (
	"context"
	"log/slog"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/cron"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/admin"
	adminInterfaces "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/admin/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/group"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CallbacksService struct {
	//More of a placeholder, which will contain inject google services to handle callbacks
	usersRepo       interfaces.UsersRepository
	lessonsRequests cron.LessonsRequestsRepository
	requests        interfaces.RequestsRepository
	cache           interfaces.HandlersCache
	adminRequests   interfaces.AdminRequestsRepository
	lessons         adminInterfaces.LessonsService
	sheets          cron.SheetsApi
	usersCron       cron.UsersRepo
	lessonsCron     cron.LessonsRepo
}

func NewCallbackService(usersRepo interfaces.UsersRepository, cache interfaces.HandlersCache, lessonsRequests cron.LessonsRequestsRepository,
	requests interfaces.RequestsRepository, adminRequests interfaces.AdminRequestsRepository,
	lessons adminInterfaces.LessonsService, sheets cron.SheetsApi, lessonsCron cron.LessonsRepo, usersCron cron.UsersRepo) *CallbacksService {
	return &CallbacksService{
		usersRepo:       usersRepo,
		cache:           cache,
		requests:        requests,
		adminRequests:   adminRequests,
		lessonsRequests: lessonsRequests,
		lessons:         lessons,
		usersCron:       usersCron,
	}
}

type CallbackHandler interface {
	HandleCallback(ctx context.Context, update *tgbotapi.Update, bot *tgutils.Bot) error
}

func (serv *CallbacksService) HandleCallbacks(update *tgbotapi.Update, bot *tgutils.Bot) {
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
	case strings.HasPrefix(update.CallbackData(), cron.REMINDER_CALLBACKS):
		callback_handler = cron.NewSheetsRefreshCallbackHandler(serv.lessonsRequests, serv.sheets, serv.usersCron, serv.lessonsCron)
	}

	err := callback_handler.HandleCallback(ctx, update, bot)
	if err != nil {
		slog.Error(err.Error())
	}
}
