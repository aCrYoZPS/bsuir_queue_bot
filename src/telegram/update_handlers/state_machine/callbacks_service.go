package stateMachine

import (
	"log/slog"
	"slices"
	"strings"

	sheetsapi "github.com/aCrYoZPS/bsuir_queue_bot/src/google_docs/sheets_api"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/admin"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)


type CallbacksService struct {
	//More of a placeholder, which will contain inject google services to handle callbacks
	sheets     sheetsapi.SheetsApi
	usersRepo  interfaces.UsersRepository
	cache      interfaces.HandlersCache
}

func NewCallbackService(usersRepo interfaces.UsersRepository, cache interfaces.HandlersCache, sheets sheetsapi.SheetsApi) *CallbacksService {
	return &CallbacksService{
		sheets:     sheets,
		usersRepo:  usersRepo,
		cache:      cache,
	}
}

type CallbackHandler interface {
	HandleCallback(update *tgbotapi.Update, bot *tgbotapi.BotAPI) error
}

func (serv *CallbacksService) HandleCallbacks(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if update.CallbackQuery == nil {
		slog.Error("no callback in update")
		return
	}

	var callback_handler CallbackHandler
	switch {
	case strings.HasPrefix(update.CallbackQuery.Data, constants.ADMIN_CALLBACKS):
		callback_handler = admin.NewAdminCallbackHandler(serv.usersRepo, serv.cache, serv.sheets)
	}
	err := callback_handler.HandleCallback(update, bot)
	if err != nil {
		slog.Error(err.Error())
	}
}

var disciplines = []string{"ООП", "АВС", "ИГИ", "МЧА"}

const CHUNK_SIZE = 4

func createDisciplinesKeyboard() *tgbotapi.InlineKeyboardMarkup {
	markup := [][]tgbotapi.InlineKeyboardButton{}
	for chunk := range slices.Chunk(disciplines, CHUNK_SIZE) {
		row := []tgbotapi.InlineKeyboardButton{}
		for _, discipline := range chunk {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(discipline, discipline))
		}
		markup = append(markup, row)
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(markup...)
	return &keyboard
}
