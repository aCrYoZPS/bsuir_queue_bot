package update_handlers

import (
	"log/slog"
	"slices"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	HELP_COMMAND      = "/help"
	SUBMIT_COMMAND    = "/submit"
	ASSIGN_COMMAND    = "/assign"
	ADD_USERS_COMMAND = "/add"
)

var userCommands = []tgbotapi.BotCommand{
	{Command: HELP_COMMAND, Description: "Команды и информация"},
	{Command: SUBMIT_COMMAND, Description: "Запись на сдачу лабораторной"},
	{Command: ASSIGN_COMMAND, Description: "Отправка заявки на роль администратора группы"},
}

func GetUserCommands() []tgbotapi.BotCommand {
	return userCommands
}

var adminCommands = []tgbotapi.BotCommand{
	{Command: ADD_USERS_COMMAND, Description: "Добавление пользователей телеграма как членов группы (необходимо для отправления заявок)"},
}

func GetAdminCommands() []tgbotapi.BotCommand {
	return adminCommands
}

type StateMachine interface {
	HandleState(chatId int64, message *tgbotapi.Message) error
}
type MessagesService struct {
	cache        interfaces.HandlersCache
	stateMachine StateMachine
}

func NewMessagesHandler(stateMachine StateMachine, cache interfaces.HandlersCache) *MessagesService {
	tgbotapi.NewSetMyCommands(userCommands...)
	return &MessagesService{cache: cache, stateMachine: stateMachine}
}

func (srv *MessagesService) HandleCommands(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if update.Message.Command() != "" {
		isCommand := func(compared string) func(command tgbotapi.BotCommand) bool {
			return func(command tgbotapi.BotCommand) bool {
				commandText, _ := strings.CutPrefix(command.Command, "/")
				return compared == commandText
			}
		}
		commands := slices.Concat(userCommands, adminCommands)
		if !slices.ContainsFunc(commands, isCommand(update.Message.Command())) {
			if _, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Незнакомая команда")); err != nil {
				slog.Error(err.Error())
			}
			return
		}
		err := srv.stateMachine.HandleState(update.Message.Chat.ID, update.Message)
		if err != nil {
			slog.Error(err.Error())
			return
		}
	}
}

func (srv *MessagesService) HandleMessages(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	err := srv.stateMachine.HandleState(update.Message.Chat.ID, update.Message)
	if err != nil {
		slog.Error(err.Error())
	}
}
