package stateMachine

import (
	"errors"
	"sync"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	SELECT_SUBJECT_STATE StateName = "subject"
	SELECT_DATE_STATE    StateName = "date"
	IDLE_STATE           StateName = ""
)

var once sync.Once

type GroupsService interface {
	DoesGroupExist(name string) (bool, error)
}

func InitStates(conf *statesConfig) {
	once.Do(
		func() {
			states = []State{}
			states = append(states,
				newIdleState(conf.cache, conf.bot), newAdminSubmitState(conf.cache, conf.bot),
				newAdminSubmittingNameState(conf.cache, conf.bot), newAdminSubmitingGroupState(conf.cache, conf.bot, conf.groupsService),
				newAdminSubmitingProofState(conf.cache, conf.bot))
		},
	)
}

var states = []State{}

func getStateByName(name string) (State, error) {
	for _, state := range states {
		if state.StateName() == StateName(name) {
			return state, nil
		}
	}
	return &idleState{}, errors.New("no such state")
}

type idleState struct {
	cache interfaces.HandlersCache
	bot   *tgbotapi.BotAPI
	State
}

func newIdleState(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI) *idleState {
	return &idleState{cache: cache, bot: bot}
}

func (state *idleState) Handle(chatId int64, message *tgbotapi.Message) error {
	switch message.Text {
	case update_handlers.ASSIGN_COMMAND:
		state.cache.SaveState(*interfaces.NewCachedInfo(chatId, string(ADMIN_SUBMIT_START_STATE)))
		state, err := getStateByName(string(ADMIN_SUBMIT_START_STATE))
		if err != nil {
			return err
		}
		return state.Handle(chatId, message)
	}
	return errors.Join(errors.ErrUnsupported, errors.New("answers are only to commands"))
}

func (*idleState) StateName() StateName {
	return IDLE_STATE
}
