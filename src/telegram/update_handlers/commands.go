package update_handlers

import (
	"log/slog"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/logging"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	HELP_COMMAND   = "help"
	SUBMIT_COMMAND = "submit"
	ASSIGN_COMMAND = "assign"
)

var commands = []tgbotapi.BotCommand{
	{Command: HELP_COMMAND, Description: "Команды и информация"},
	{Command: SUBMIT_COMMAND, Description: "Запись на сдачу лабораторной"},
	{Command: ASSIGN_COMMAND, Description: "Отправка заявки на роль администратора группы"},
}

type StateMachine interface {
	HandleState(chatId int64, message string) error
}
type MessagesService struct {
	cache        interfaces.HandlersCache
	stateMachine StateMachine
}

func NewMessagesHandler(cache interfaces.HandlersCache) *MessagesService {
	tgbotapi.NewSetMyCommands(commands...)
	return &MessagesService{cache: cache}
}

func (srv *MessagesService) HandleCommands(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
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
			slog.Error(err.Error())
		}
	default:
		if _, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Незнакомая команда")); err != nil {
			slog.Error(err.Error())
		}
	}
}

func (srv *MessagesService) HandleMessages(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if update.Message.Text == "" {
		text := "Enter a valid text message please"
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
		_, err := bot.Send(msg)
		if err != nil {
			slog.Error(err.Error())
		}
	} else {
		err := srv.stateMachine.HandleState(update.Message.Chat.ID, update.Message.Text)
		if err != nil {
			slog.Error(err.Error())
		}
	}
}
