package stateMachine

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type idleState struct {
	cache     interfaces.HandlersCache
	bot       *tgutils.Bot
	usersRepo interfaces.UsersRepository
	State
}

func newIdleState(cache interfaces.HandlersCache, bot *tgutils.Bot, usersRepo interfaces.UsersRepository) *idleState {
	return &idleState{cache: cache, bot: bot, usersRepo: usersRepo}
}

func (state *idleState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	var currentState State
	switch message.Text {
	case update_handlers.ASSIGN_COMMAND:
		err := state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.ADMIN_SUBMIT_START_STATE))
		if err != nil {
			return fmt.Errorf("failed to transition from idle to admin submit state: %w", err)
		}
		currentState = getStateByName(constants.ADMIN_SUBMIT_START_STATE)
		if currentState == nil {
			return fmt.Errorf("failed to get state for name %s", currentState)
		}
		err = currentState.Handle(ctx, message)
		if err != nil {
			return err
		}
	case update_handlers.HELP_COMMAND:
		var commands []tgbotapi.BotCommand
		commands = append(commands, slices.Concat(update_handlers.GetUserCommands(), update_handlers.GetAdminCommands())...)
		builder := strings.Builder{}
		for _, command := range commands {
			builder.WriteString(command.Command)
			builder.WriteString(" - ")
			builder.WriteString(command.Description)
			builder.WriteByte('\n')
		}
		_, err := state.bot.SendCtx(ctx, tgbotapi.NewMessage(message.Chat.ID, builder.String()))
		if err != nil {
			return fmt.Errorf("failed to send message during help command: %w", err)
		}
	case update_handlers.JOIN_GROUP_COMMAND:
		err := state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.GROUP_SUBMIT_START_STATE))
		if err != nil {
			return err
		}
		currentState = getStateByName(constants.GROUP_SUBMIT_START_STATE)
		if currentState == nil {
			return fmt.Errorf("couldn't find state for %s command", update_handlers.JOIN_GROUP_COMMAND)
		}
	case update_handlers.SUBMIT_COMMAND:
		err := state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.LABWORK_SUBMIT_START_STATE))
		if err != nil {
			return fmt.Errorf("failed to transition from idle state to labwork submit state")
		}
		currentState = getStateByName(constants.LABWORK_SUBMIT_START_STATE)
		if currentState == nil {
			return fmt.Errorf("couldn't find state for %s command", constants.LABWORK_SUBMIT_START_STATE)
		}
	case update_handlers.ADD_LABWORK_COMMAND:
		err := state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.LABWORK_ADD_START_STATE))
		if err != nil {
			return fmt.Errorf("failed to transition from idle state to labwork add state")
		}
		currentState = getStateByName(constants.LABWORK_ADD_START_STATE)
		if currentState == nil {
			return fmt.Errorf("couldn't find state for %s command", constants.LABWORK_ADD_START_STATE)
		}
	default:
		return errors.Join(errors.ErrUnsupported, errors.New("answers are only to commands"))
	}
	err := currentState.Handle(ctx, message)
	if err != nil {
		return err
	}
	return nil
}

func (*idleState) StateName() string {
	return constants.IDLE_STATE
}
