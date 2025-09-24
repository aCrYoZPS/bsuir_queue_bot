package stateMachine

import (
	"context"
	"slices"
	"sync"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/admin"
	customlabworks "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/custom_labworks"
	groups "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/group"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/labworks"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type State interface {
	StateName() string
	Handle(ctx context.Context, message *tgbotapi.Message) error
	Revert(ctx context.Context, message *tgbotapi.Message) error
}

var once sync.Once

func InitStates(conf *statesConfig) {
	once.Do(
		func() {
			states = []State{newIdleState(conf.cache, conf.bot, conf.usersRepo, conf.groupsRepo)}
			states = slices.Concat(states, createAdminStates(conf), createGroupStates(conf), createLabworksStates(conf), createCustomLabworkStates(conf))
		},
	)
}

func createAdminStates(conf *statesConfig) []State {
	return []State{admin.NewAdminSubmitState(conf.cache, conf.bot, conf.usersRepo),
		admin.NewAdminSubmittingNameState(conf.cache, conf.bot, conf.machine), admin.NewAdminSubmitingGroupState(conf.cache, conf.bot, conf.groupsRepo, conf.machine),
		admin.NewAdminSubmitingProofState(conf.cache, conf.bot, conf.adminRequests, conf.machine), admin.NewAdminWaitingProofState(conf.cache, conf.bot)}
}

func createGroupStates(conf *statesConfig) []State {
	return []State{groups.NewGroupSubmitState(conf.cache, conf.bot, conf.groupsRepo, conf.usersRepo), groups.NewGroupSubmitNameState(conf.cache, conf.bot, conf.groupsRepo, conf.requests),
		groups.NewGroupSubmitGroupNameState(conf.cache, conf.bot, conf.groupsRepo), groups.NewGroupWaitingState(conf.cache, conf.bot)}
}

func createLabworksStates(conf *statesConfig) []State {
	return []State{labworks.NewLabworkSubmitProofState(conf.bot, conf.cache, conf.groupsRepo, conf.requests), labworks.NewLabworkSubmitNumberState(conf.bot, conf.cache, conf.labworks, conf.usersRepo, conf.machine),
		labworks.NewLabworkSubmitStartState(conf.bot, conf.cache, conf.labworks, conf.usersRepo), labworks.NewLabworkSubmitWaitingState(conf.bot, conf.cache, conf.machine)}
}

func createCustomLabworkStates(conf *statesConfig) []State {
	return []State{customlabworks.NewLabworkAddStartState(conf.bot, conf.cache, conf.usersRepo), customlabworks.NewLabworkAddSubmitNameState(conf.bot, conf.cache),
		customlabworks.NewLabworkAddWaitingState(conf.bot)}
}

var states = []State{}

func getStateByName(name string) State {
	for _, state := range states {
		if state.StateName() == (name) {
			return state
		}
	}
	return nil
}
