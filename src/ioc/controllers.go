package ioc

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/cron"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/cron/schedule"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/logging"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/bot"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	"github.com/go-co-op/gocron/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var useTgBot = provider(
	func() *tgutils.Bot {
		bot_token := os.Getenv("BOT_TOKEN")
		debug := os.Getenv("DEBUG")
		bot, err := tgbotapi.NewBotAPI(bot_token)
		if err != nil {
			logging.FatalLog(err.Error())
		}
		if strings.EqualFold(debug, "true") {
			bot.Debug = true
		}
		return tgutils.NewBot(bot)
	},
)

var UseBotController = provider(
	func() *bot.BotController {
		bot, err := bot.NewBotController(useTgBot(), UseMessageService(), UseCallbacksService())
		if err != nil {
			logging.FatalLog(err.Error())
		}
		RegisterRoutes(useMux())
		return bot
	},
)

var useMux = provider(
	func() *tgutils.Mux {
		mux := tgutils.NewMux(useHandlersCache(), useTgBot())
		return mux
	},
)

var UseTasksController = provider(
	func() *cron.TasksController {
		controller := cron.NewTasksController(useTasksRepository())
		AddClearTask(controller)
		AddRefreshTask(controller)
		AddReminderTask(controller)
		return controller
	},
)

var daily = gocron.CronJob("00 22 * * *", false)

func AddReminderTask(controller *cron.TasksController) {
	reminder := cron.NewReminderTask(UseSheetsApiService(), useLessonsRepository(), useLessonsRequestsRepository(),
		useUsersRepository(), useTgBot())
	controller.AddTask(daily, gocron.NewTask(func(ctx context.Context) {
		const sheetsRefreshTimeout = 5 * time.Minute
		ctx, cancel := context.WithTimeout(ctx, sheetsRefreshTimeout)
		defer cancel()
		reminder.Run(ctx)
	}), gocron.WithName("sheets refresh"))
}

func AddClearTask(controller *cron.TasksController) {
	clear := cron.NewClearLessonsTask(UseSheetsApiService(), useLessonsRepository(), UseDriveApiService())
	controller.AddTask(daily, gocron.NewTask(func(ctx context.Context) {
		const sheetsClearTimeout = 5 * time.Minute
		ctx, cancel := context.WithTimeout(ctx, sheetsClearTimeout)
		defer cancel()
		clear.Run(ctx)
	}), gocron.WithName("sheets clear"))
}

func AddRefreshTask(controller *cron.TasksController) {
	refresh := schedule.NewRefreshScheduleTask(useGroupsRepository(), UseLessonsService())
	controller.AddTask(daily, gocron.NewTask(func(ctx context.Context) {
		const sheetsClearTimeout = 5 * time.Minute
		ctx, cancel := context.WithTimeout(ctx, sheetsClearTimeout)
		defer cancel()
		refresh.Run(ctx)
	}), gocron.WithName("schedule refresh"))
}
