package ioc

import (
	"os"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/cron"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/logging"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/bot"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
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
		return cron.NewTasksController(UseSheetsApiService(), useLessonsRepository(),
			useLessonsRequestsRepository(), useUsersRepository(),
			UseDriveApiService(), useTasksRepository(), useTgBot())
	},
)
