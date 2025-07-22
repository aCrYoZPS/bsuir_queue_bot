package stateMachine

import (
	"errors"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type idleState struct {
	cache     interfaces.HandlersCache
	bot       *tgbotapi.BotAPI
	usersRepo interfaces.UsersRepository
	State
}

func newIdleState(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI, usersRepo interfaces.UsersRepository) *idleState {
	return &idleState{cache: cache, bot: bot, usersRepo: usersRepo}
}

func (state *idleState) Handle(chatId int64, message *tgbotapi.Message) error {
	var currentState State
	switch message.Text {
	case update_handlers.ASSIGN_COMMAND:
		err := state.cache.SaveState(*interfaces.NewCachedInfo(chatId, constants.ADMIN_SUBMIT_START_STATE))
		if err != nil {
			return err
		}
		currentState, err = getStateByName(constants.ADMIN_SUBMIT_START_STATE)
		if err != nil {
			return err
		}
		err = currentState.Handle(chatId, message)
		if err != nil {
			return err
		}
	case update_handlers.HELP_COMMAND:
		var commands []tgbotapi.BotCommand
		commands = append(commands, update_handlers.GetUserCommands()...)
		builder := strings.Builder{}
		for _, command := range commands {
			builder.WriteString(command.Command)
			builder.WriteString(" - ")
			builder.WriteString(command.Description)
			builder.WriteByte('\n')
		}
		_, err := state.bot.Send(tgbotapi.NewMessage(chatId, builder.String()))
		if err != nil {
			return err
		}
	case update_handlers.JOIN_GROUP_COMMAND:
		err := state.cache.SaveState(*interfaces.NewCachedInfo(chatId, constants.GROUP_SUBMIT_START_STATE))
		if err != nil {
			return err
		}
		currentState, err = getStateByName(constants.GROUP_SUBMIT_START_STATE)
		if err != nil {
			return err
		}
		err = currentState.Handle(chatId, message)
		if err != nil {
			return err
		}
	default:
		return errors.Join(errors.ErrUnsupported, errors.New("answers are only to commands"))
	}
	return nil
}

func (*idleState) StateName() string {
	return constants.IDLE_STATE
}
