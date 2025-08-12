package stateMachine

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/cron"
	sheetsapi "github.com/aCrYoZPS/bsuir_queue_bot/src/google_docs/sheets_api"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/admin"
	adminInterfaces "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/admin/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	customlabworks "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/custom_labworks"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/group"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/labworks"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type LessonsService interface {
	cron.LessonRepo
	adminInterfaces.LessonsService
	labworks.LabworksService
	interfaces.LessonsRepository
}

type UsersRepository interface {
	interfaces.UsersRepository
	cron.UsersRepoReminder
}

type LessonsRequestsService interface {
	cron.LessonsRequestRepo
	cron.LessonsRequestsRepositoryReminder
	interfaces.LessonsRequestsRepository
}

type SheetsApi interface {
	cron.SheetsApi
	cron.SheetsApiReminder
	cron.SheetsApiClear
	sheetsapi.SheetsApi
}

type CallbacksService struct {
	//More of a placeholder, which will contain inject google services to handle callbacks
	usersRepo       UsersRepository
	lessonsRequests LessonsRequestsService
	requests        interfaces.RequestsRepository
	cache           interfaces.HandlersCache
	adminRequests   interfaces.AdminRequestsRepository
	lessons         LessonsService
	sheets          SheetsApi
}

func NewCallbackService(usersRepo UsersRepository, cache interfaces.HandlersCache, lessonsRequests LessonsRequestsService,
	requests interfaces.RequestsRepository, adminRequests interfaces.AdminRequestsRepository,
	lessons LessonsService, sheets SheetsApi, lessonsCron cron.LessonsRepoReminder) *CallbacksService {
	return &CallbacksService{
		sheets:          sheets,
		usersRepo:       usersRepo,
		cache:           cache,
		requests:        requests,
		adminRequests:   adminRequests,
		lessonsRequests: lessonsRequests,
		lessons:         lessons,
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
	case strings.HasPrefix(update.CallbackData(), constants.LABWORK_CALLBACKS):
		callback_handler = labworks.NewLabworksCallbackHandler(bot, serv.cache, serv.lessons, serv.requests, serv.lessonsRequests, serv.usersRepo, serv.sheets)
	case strings.HasPrefix(update.CallbackData(), cron.REMINDER_CALLBACKS):
		callback_handler = cron.NewSheetsRefreshCallbackHandler(serv.lessonsRequests, serv.sheets, serv.usersRepo, serv.lessons)
	case strings.HasPrefix(update.CallbackData(), constants.CALENDAR_CALLBACKS):
		callback_handler = customlabworks.NewCalendarCallbackHandler(bot, serv.cache)
	case strings.HasPrefix(update.CallbackData(), constants.TIME_PICKER_CALLBACKS):
		callback_handler = customlabworks.NewTimePickerCallbackHandler(bot, serv.lessons, serv.cache)
	case strings.HasPrefix(update.CallbackData(), constants.IGNORE_CALLBACKS):
		callbackConfig := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		_, err := bot.Request(callbackConfig)
		if err != nil {
			slog.Error(fmt.Errorf("failed to send empty callback during ignore callback handling: %w", err).Error())
		}
		return
	}

	err := callback_handler.HandleCallback(ctx, update, bot)
	if err != nil {
		slog.Error(err.Error())
	}
}
