package stateMachine

import (
	"context"
	"fmt"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/labworks"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type StateName string

type StateMachine struct {
	update_handlers.StateMachine
	cache     interfaces.HandlersCache
	bot       *tgutils.Bot
	usersRepo interfaces.UsersRepository
}

type statesConfig struct {
	machine          *StateMachine
	cache            interfaces.HandlersCache
	bot              *tgutils.Bot
	groupsRepo       interfaces.GroupsRepository
	usersRepo        interfaces.UsersRepository
	requests         interfaces.RequestsRepository
	adminRequests    interfaces.AdminRequestsRepository
	labworks         labworks.LabworksService
	labworksRequests interfaces.LessonsRequestsRepository
}

func NewStatesConfig(state *StateMachine, cache interfaces.HandlersCache, bot *tgutils.Bot, groupsRepo interfaces.GroupsRepository, usersRepo interfaces.UsersRepository,
	requests interfaces.RequestsRepository, adminRequests interfaces.AdminRequestsRepository, labworks labworks.LabworksService, labworksRequests interfaces.LessonsRequestsRepository) *statesConfig {
	return &statesConfig{
		machine:          state,
		cache:            cache,
		bot:              bot,
		groupsRepo:       groupsRepo,
		usersRepo:        usersRepo,
		requests:         requests,
		labworks:         labworks,
		adminRequests:    adminRequests,
		labworksRequests: labworksRequests,
	}
}

func NewStateMachine(conf *statesConfig) *StateMachine {
	machine := &StateMachine{cache: conf.cache, bot: conf.bot, usersRepo: conf.usersRepo}
	conf.machine = machine
	InitStates(conf)
	return machine
}

func (machine *StateMachine) HandleStateMu(ctx context.Context, message *tgbotapi.Message) error {
	mu := machine.cache.AcquireLock(ctx, message.Chat.ID)
	mu.Lock()

	defer mu.Unlock()
	defer machine.cache.ReleaseLock(ctx, message.Chat.ID)

	info, err := machine.cache.GetState(ctx, message.Chat.ID)
	if err != nil {
		return fmt.Errorf("couldn't get state in state machine: %w", err)
	}

	var state State
	if info != nil {
		state = getStateByName(info.State())
		if state == nil {
			return fmt.Errorf("failed to get state for name %s", info.State())
		}
	} else {
		state = getStateByName(constants.IDLE_STATE)
		if state == nil {
			return fmt.Errorf("failed to get idle state for name %s", info.State())
		}
	}
	if message.Command() == strings.Trim(update_handlers.REVERT_COMMAND, "/") {
		return state.Revert(ctx, message)
	}
	return state.Handle(ctx, message)
}

// Only for recursion calls, when we know that state is concurrently safe!
func (machine *StateMachine) HandleState(ctx context.Context, message *tgbotapi.Message) error {
	info, err := machine.cache.GetState(ctx, message.Chat.ID)
	if err != nil {
		return fmt.Errorf("couldn't get state in state machine: %w", err)
	}

	var state State
	if info != nil {
		state = getStateByName(info.State())
		if state == nil {
			return fmt.Errorf("failed to get state for name %s", info.State())
		}
	} else {
		state = getStateByName(constants.IDLE_STATE)
		if state == nil {
			return fmt.Errorf("failed to get idle state for name %s", info.State())
		}
	}
	if message.Command() == strings.Trim(update_handlers.REVERT_COMMAND, "/") {
		return state.Revert(ctx, message)
	}
	return state.Handle(ctx, message)
}
