package update_handlers

import (
	"context"
	"log/slog"
	"runtime/debug"
	"slices"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/constants"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var userCommands = []tgbotapi.BotCommand{
	{Command: constants.HELP_COMMAND, Description: "Команды и информация"},
	{Command: constants.SUBMIT_COMMAND, Description: "Запись на сдачу лабораторной"},
	{Command: constants.ASSIGN_COMMAND, Description: "Отправка заявки на роль администратора группы"},
	{Command: constants.JOIN_GROUP_COMMAND, Description: "Отправка заявки на участие в группе"},
	{Command: constants.QUEUE_COMMAND, Description: "Получение очереди своей группы"},
	{Command: constants.REVERT_COMMAND, Description: "Откат к предыдущему состоянию"},
	{Command: constants.TABLE_COMMAND, Description: "Получение ссылки на гугл-таблицу своей группы"},
}

var adminCommands = []tgbotapi.BotCommand{
	{Command: constants.ADD_LABWORK_COMMAND, Description: "Добавление собственной пары"},
}

func GetUserCommands() []tgbotapi.BotCommand {
	return userCommands
}

func GetAdminCommands() []tgbotapi.BotCommand {
	return adminCommands
}

type StateMachine interface {
	Handle(ctx context.Context, message *tgbotapi.Message) error
}
type MessagesService struct {
	cache        interfaces.HandlersCache
	stateMachine StateMachine
}

func NewMessagesHandler(stateMachine StateMachine, cache interfaces.HandlersCache) *MessagesService {
	tgbotapi.NewSetMyCommands(slices.Concat(userCommands, adminCommands)...)
	return &MessagesService{cache: cache, stateMachine: stateMachine}
}

func (srv *MessagesService) HandleMessages(update *tgbotapi.Update, bot *tgutils.Bot) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Recovered from panic", "err", r)
			debug.PrintStack()
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), constants.DEFAULT_TIMEOUT)
	defer cancel()

	mu := srv.cache.AcquireLock(ctx, update.Message.Chat.ID, update.CallbackData())
	mu.Lock()

	defer mu.Unlock()
	defer srv.cache.ReleaseLock(ctx, update.Message.Chat.ID, update.CallbackData())

	err := srv.stateMachine.Handle(ctx, update.Message)
	if err != nil {
		slog.Error(err.Error())
	}
}
