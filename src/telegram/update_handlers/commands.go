package update_handlers

import (
<<<<<<< HEAD
	"github.com/aCrYoZPS/bsuir_queue_bot/src/logging"
=======
	"log/slog"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
>>>>>>> b9dfac65e2826f04e755903fa773ab7503d2fb20
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	HELP_COMMAND   = "help"
	SUBMIT_COMMAND = "submit"
	ASSIGN_COMMAND = "assign"

	SELECT_SUBJECT_STATE = "subject"
	SELECT_DATE_STATE    = "date"
	SUBMIT_PROOF_STATE   = "proof"
	IDLE_STATE           = ""
)

var commands = []tgbotapi.BotCommand{
	{Command: HELP_COMMAND, Description: "Команды и информация"},
	{Command: SUBMIT_COMMAND, Description: "Запись на сдачу лабораторной"},
	{Command: ASSIGN_COMMAND, Description: "Отправка заявки на роль администратора группы"},
}

type MessagesService struct {
	cache interfaces.HandlersCache
}

func NewMessagesHandler(cache interfaces.HandlersCache) *MessagesService {
	tgbotapi.NewSetMyCommands(commands...)
	return &MessagesService{cache: cache}
}

func (*MessagesService) HandleCommands(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	switch update.Message.Command() {
	case HELP_COMMAND:
		var text string
		for _, command := range commands {
			text += "/" + command.Command + ": " + command.Description + "\n"
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
		_, err := bot.Send(msg)
		if err != nil {
			logging.Error(err.Error())
		}
	case SUBMIT_COMMAND:
		text := "Выберите предмет"
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
		msg.ReplyMarkup = createDisciplinesKeyboard()
		_, err := bot.Send(msg)
		if err != nil {
			logging.Error(err.Error())
		}
	case ASSIGN_COMMAND:
		text := "Введите номер группы и отправьте подтверждение принадлежности к ней"
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
		_, err := bot.Send(msg)
		if err != nil {
			logging.Error(err.Error())
		}
	default:
		if _, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Незнакомая команда")); err != nil {
			logging.Error(err.Error())
		}
	}
}

func HandleMessage(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	chatId := update.Message.Chat.ID
	state := userStates[chatId]

	switch state {
	case IDLE_STATE:
		return
	case SELECT_DATE_STATE:
	}
}

// Probably is not worthy enough to be a separate package at all, but for the readability, why not
func InitCommands() {
	tgbotapi.NewSetMyCommands(
		commands...,
	)
}
