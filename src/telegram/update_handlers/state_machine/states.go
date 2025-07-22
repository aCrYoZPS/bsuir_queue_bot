package stateMachine

import (
	"errors"
	"sync"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/admin"
	groups "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/group"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type State interface {
	StateName() string
	Handle(chatId int64, message *tgbotapi.Message) error
}

var once sync.Once

func InitStates(conf *statesConfig) {
	once.Do(
		func() {
			states = []State{}
			states = append(states,
				newIdleState(conf.cache, conf.bot, conf.usersRepo), admin.NewAdminSubmitState(conf.cache, conf.bot),
				admin.NewAdminSubmittingNameState(conf.cache, conf.bot), admin.NewAdminSubmitingGroupState(conf.cache, conf.bot, conf.groupsRepo),
				admin.NewAdminSubmitingProofState(conf.cache, conf.bot), admin.NewAdminWaitingProofState(conf.cache, conf.bot),
				groups.NewGroupSubmitState(conf.cache, conf.bot, conf.groupsRepo, conf.usersRepo), groups.NewGroupSubmitNameState(conf.cache, conf.bot, conf.groupsRepo, conf.requests),
				groups.NewGroupWaitingState(conf.cache, conf.bot))
		},
	)
}

var states = []State{}

func getStateByName(name string) (State, error) {
	for _, state := range states {
		if state.StateName() == (name) {
			return state, nil
		}
	}
	return &idleState{}, errors.New("no such state")
}
