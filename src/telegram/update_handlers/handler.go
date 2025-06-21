package update_handlers

import (
	"log/slog"
	"slices"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleCommands(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	switch update.Message.Command() {
	case HELP_COMMAND:
		var text string
		for _, command := range commands {
			text += "/" + command.Command + ": " + command.Description + "\n"
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
		_, err := bot.Send(msg)
		if err != nil {
			slog.Error(err.Error())
		}
	case SUBMIT_COMMAND:
		text := "Select preferred discipline"
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
		msg.ReplyMarkup = createDisciplinesKeyboard()
		_, err := bot.Send(msg)
		if err != nil {
			slog.Error(err.Error())
		}
	default:
		if _, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command")); err != nil {
			slog.Error(err.Error())
		}
	}
}

func HandleCallbacks(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	callback := tgbotapi.NewCallback(update.CallbackQuery.InlineMessageID, update.CallbackQuery.Data)
	if _, err := bot.Request(callback); err != nil {
		slog.Error(err.Error())
	}
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data)
	if _, err := bot.Send(msg); err != nil {
		slog.Error(err.Error())
	}
}

var disciplines = []string{"ООП", "АВС", "ИГИ", "МЧА"}

const chunk_size = 4

func createDisciplinesKeyboard() *tgbotapi.InlineKeyboardMarkup {
	markup := [][]tgbotapi.InlineKeyboardButton{}
	for chunk := range slices.Chunk(disciplines, chunk_size) {
		row := []tgbotapi.InlineKeyboardButton{}
		for _, discipline := range chunk {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(discipline, discipline))
		}
		markup = append(markup, row)
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(markup...)
	return &keyboard
}
