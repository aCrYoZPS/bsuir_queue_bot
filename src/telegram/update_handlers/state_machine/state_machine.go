package stateMachine

import (
	"context"
	"fmt"

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
	cache         interfaces.HandlersCache
	bot           *tgutils.Bot
	groupsRepo    interfaces.GroupsRepository
	usersRepo     interfaces.UsersRepository
	requests      interfaces.RequestsRepository
	adminRequests interfaces.AdminRequestsRepository
	labworks      labworks.LabworksService
}

func NewStatesConfig(cache interfaces.HandlersCache, bot *tgutils.Bot, groupsRepo interfaces.GroupsRepository, usersRepo interfaces.UsersRepository,
	requests interfaces.RequestsRepository, adminRequests interfaces.AdminRequestsRepository, labworks labworks.LabworksService) *statesConfig {
	return &statesConfig{
		cache:         cache,
		bot:           bot,
		groupsRepo:    groupsRepo,
		usersRepo:     usersRepo,
		requests:      requests,
		labworks:      labworks,
		adminRequests: adminRequests,
	}
}

func NewStateMachine(conf *statesConfig) *StateMachine {
	InitStates(conf)
	return &StateMachine{cache: conf.cache, bot: conf.bot, usersRepo: conf.usersRepo}
}

func (machine *StateMachine) HandleState(ctx context.Context, message *tgbotapi.Message) error {
	mu := machine.cache.AcquireLock(ctx, message.Chat.ID)
	mu.Lock()

	defer mu.Unlock()
	defer machine.cache.ReleaseLock(ctx, message.Chat.ID)

	info, err := machine.cache.GetState(ctx, message.Chat.ID)
	if err != nil {
		return fmt.Errorf("couldn't get state in state machine: %w",err)
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
	return state.Handle(ctx, message)
}
