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

	SELECT_SUBJECT_STATE = "subject"
	SELECT_DATE_STATE    = "date"
	SUBMIT_PROOF_STATE   = "proof"
	IDLE_STATE           = ""
)

var commands = []tgbotapi.BotCommand{
	{Command: HELP_COMMAND, Description: "Commands and their short description"},
	{Command: SUBMIT_COMMAND, Description: "Yeah, I am lazy even for that"},
	{Command: ASSIGN_COMMAND, Description: "Submit your request for becoming admin of the chosen group"},
}

type StateMachine interface {
	HandleState(curState, message string) error
}
type MessagesService struct {
	cache        interfaces.HandlersCache
	stateMachine StateMachine
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
		text := "Enter the desired group number and send proofs of belonging to it"
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
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

func (srv *MessagesService) HandleMessages(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if update.Message.Text == "" {
		text := "Enter a valid text message please"
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
		_, err := bot.Send(msg)
		if err != nil {
			slog.Error(err.Error())
		}
	} else {
		info, err := srv.cache.Get(update.Message.Chat.ID)
		if err != nil {
			slog.Error(err.Error())
		}
		err = srv.stateMachine.HandleState(info.State(), update.Message.Text)
		if err != nil {
			slog.Error(err.Error())
		}
	}
}
