package update_handlers

import (
	"context"
	"log/slog"
	"slices"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	HELP_COMMAND        = "/help"
	SUBMIT_COMMAND      = "/submit"
	ASSIGN_COMMAND      = "/assign"
	JOIN_GROUP_COMMAND  = "/join"
	ADD_LABWORK_COMMAND = "/add"
)

var userCommands = []tgbotapi.BotCommand{
	{Command: HELP_COMMAND, Description: "Команды и информация"},
	{Command: SUBMIT_COMMAND, Description: "Запись на сдачу лабораторной"},
	{Command: ASSIGN_COMMAND, Description: "Отправка заявки на роль администратора группы"},
	{Command: JOIN_GROUP_COMMAND, Description: "Отправка заявки на участие в группе"},
}

var adminCommands = []tgbotapi.BotCommand{
	{Command: ADD_LABWORK_COMMAND, Description: "Добавление собственной пары"},
}

func GetUserCommands() []tgbotapi.BotCommand {
	return userCommands
}

func GetAdminCommands() []tgbotapi.BotCommand {
	return adminCommands
}

type StateMachine interface {
	HandleState(ctx context.Context, message *tgbotapi.Message) error
}
type MessagesService struct {
	cache        interfaces.HandlersCache
	stateMachine StateMachine
}

func NewMessagesHandler(stateMachine StateMachine, cache interfaces.HandlersCache) *MessagesService {
	tgbotapi.NewSetMyCommands(userCommands...)
	return &MessagesService{cache: cache, stateMachine: stateMachine}
}

func (srv *MessagesService) HandleCommands(update *tgbotapi.Update, bot *tgutils.Bot) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.DEFAULT_TIMEOUT)
	defer cancel()
	if update.Message.Command() != "" {
		isCommand := func(compared string) func(command tgbotapi.BotCommand) bool {
			return func(command tgbotapi.BotCommand) bool {
				commandText, _ := strings.CutPrefix(command.Command, "/")
				return compared == commandText
			}
		}
		if !slices.ContainsFunc(slices.Concat(userCommands, adminCommands), isCommand(update.Message.Command())) {
			if _, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Незнакомая команда")); err != nil {
				slog.Error(err.Error())
			}
			return
		}
		err := srv.stateMachine.HandleState(ctx, update.Message)
		if err != nil {
			slog.Error(err.Error())
			return
		}
	}
}

func (srv *MessagesService) HandleMessages(update *tgbotapi.Update, bot *tgutils.Bot) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.DEFAULT_TIMEOUT)
	defer cancel()
	err := srv.stateMachine.HandleState(ctx, update.Message)
	if err != nil {
		slog.Error(err.Error())
	}
}
