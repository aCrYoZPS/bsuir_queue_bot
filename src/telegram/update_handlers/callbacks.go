package update_handlers

import (
	"slices"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/logging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CallbacksService struct {
	//More of a placeholder, which will contain inject google services to handle callbacks
}

func NewCallbackService() *CallbacksService {
	return &CallbacksService{}
}

func (*CallbacksService) HandleCallbacks(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
	if _, err := bot.Request(callback); err != nil {
		logging.Error(err.Error())
	}
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data)
	if _, err := bot.Send(msg); err != nil {
		logging.Error(err.Error())
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
