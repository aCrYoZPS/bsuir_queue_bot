package update_handlers

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/constants"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var _ (State) = (*idleState)(nil)

type idleState struct {
	cache      interfaces.HandlersCache
	bot        *tgutils.Bot
	usersRepo  interfaces.UsersRepository
	groupsRepo interfaces.GroupsRepository
	lessons    interfaces.LessonsRepository
	mux        tgutils.MuxHandler
}

func NewIdleState(cache interfaces.HandlersCache, bot *tgutils.Bot, usersRepo interfaces.UsersRepository, groupsRepo interfaces.GroupsRepository, lessons interfaces.LessonsRepository, mux tgutils.MuxHandler) *idleState {
	return &idleState{cache: cache, bot: bot, usersRepo: usersRepo, groupsRepo: groupsRepo, lessons: lessons, mux: mux}
}

func (state *idleState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	switch message.Text {
	case constants.ASSIGN_COMMAND:
		err := state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.ADMIN_SUBMIT_START_STATE))
		if err != nil {
			return fmt.Errorf("failed to transition from idle to admin submit state: %w", err)
		}
	case constants.HELP_COMMAND:
		var commands []tgbotapi.BotCommand
		commands = append(commands, GetUserCommands()...)
		user, err := state.usersRepo.GetByTgId(ctx, message.From.ID)
		if err != nil {
			return fmt.Errorf("failed to get user by id during handling help command: %w", err)
		}
		if slices.Contains(user.Roles, entities.Admin) {
			commands = append(commands, GetAdminCommands()...)
		}
		builder := strings.Builder{}
		for _, command := range commands {
			builder.WriteString(command.Command)
			builder.WriteString(" - ")
			builder.WriteString(command.Description)
			builder.WriteByte('\n')
		}
		_, err = state.bot.SendCtx(ctx, tgbotapi.NewMessage(message.Chat.ID, builder.String()))
		if err != nil {
			return fmt.Errorf("failed to send message during help command: %w", err)
		}
		return nil
	case constants.QUEUE_COMMAND:
		err := state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.QUEUE_START_STATE))
		if err != nil {
			return err
		}
	case constants.TABLE_COMMAND:
		return state.HandleTableCommand(ctx, message)
	case constants.JOIN_GROUP_COMMAND:
		err := state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.GROUP_SUBMIT_START_STATE))
		if err != nil {
			return fmt.Errorf("failed to save group submit start state: %w", err)
		}
	case constants.SUBMIT_COMMAND:
		err := state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.LABWORK_SUBMIT_START_STATE))
		if err != nil {
			return fmt.Errorf("failed to transition from idle state to labwork submit state: %w", err)
		}
	case constants.ADD_LABWORK_COMMAND:
		err := state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.LABWORK_ADD_START_STATE))
		if err != nil {
			return fmt.Errorf("failed to transition from idle state to labwork add state: %w", err)
		}
	case constants.START_COMMAND:
		user, err := state.usersRepo.GetByTgId(ctx, message.From.ID)
		if err != nil {
			return fmt.Errorf("failed to get user by id during handling start command: %w", err)
		}
		msg := tgbotapi.NewMessage(message.Chat.ID, `Воспользуйтесь /help для получения списка команд. Для отправки заявок на лабораторные
		 вы должны либо стать админом группы,с одобрения владельца бота,либо же членом группы,если у неё уже есть админ.`)
		err = tgutils.CreateStartReplyMarkup(ctx, &msg, user, state.bot)
		if err != nil {
			return fmt.Errorf("failed to create start reply markup during start command: %w", err)
		}
		return nil
	default:
		return errors.Join(errors.ErrUnsupported, errors.New("answers are only to commands"))
	}
	err := state.mux.Handle(ctx, message)
	if err != nil {
		return err
	}
	return nil
}

func (*idleState) StateName() string {
	return constants.IDLE_STATE
}

func (state *idleState) HandleTableCommand(ctx context.Context, msg *tgbotapi.Message) error {
	usr, err := state.usersRepo.GetByTgId(ctx, msg.From.ID)
	if err != nil {
		return fmt.Errorf("failed to get user by tg id during queue command handling: %w", err)
	}
	if usr.GroupId == 0 {
		_, err = state.bot.SendCtx(ctx, tgbotapi.NewMessage(msg.Chat.ID, "Вы пока не принадлежите ни к одной группе"))
		if err != nil {
			return fmt.Errorf("failed to send no group message during queue command handling: %w", err)
		}
		return nil
	}
	group, err := state.groupsRepo.GetById(ctx, int(usr.GroupId))
	if err != nil {
		return fmt.Errorf("failed to get group by id during queue command handling: %w", err)
	}
	_, err = state.bot.SendCtx(ctx, tgbotapi.NewMessage(msg.Chat.ID, state.createSheetUrl(group.SpreadsheetId)))
	if err != nil {
		return fmt.Errorf("failed to send spreadsheet url during queue command handling: %w", err)
	}
	return nil
}

func (state *idleState) createSheetUrl(spreadsheetId string) string {
	return fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/edit#gid=0", spreadsheetId)
}

func (state *idleState) Revert(ctx context.Context, msg *tgbotapi.Message) error {
	return nil
}
