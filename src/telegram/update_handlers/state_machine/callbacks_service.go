package stateMachine

import (
	"slices"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/logging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	ADMIN_CALLBACKS = "admin"
)

type CallbacksService struct {
	//More of a placeholder, which will contain inject google services to handle callbacks
}

func NewCallbackService() *CallbacksService {
	return &CallbacksService{}
}

type CallbackHandler interface {
	HandleCallback(update *tgbotapi.Update, bot *tgbotapi.BotAPI) error
}

func (*CallbacksService) HandleCallbacks(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if update.CallbackQuery == nil {
		logging.Error("no callback in update")
		return
	}

	switch update.CallbackQuery.ID {
	case ADMIN_CALLBACKS:
		//TODO:
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
