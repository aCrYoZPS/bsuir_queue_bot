package stateMachine

import (
	"errors"
	"sync"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	SELECT_SUBJECT_STATE StateName = "subject"
	SELECT_DATE_STATE    StateName = "date"
	SUBMIT_PROOF_STATE   StateName = "proof"
	ADMIN_SUBMIT_STATE   StateName = "submit"
	IDLE_STATE           StateName = ""
)

var once sync.Once

func InitStates(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI) {
	once.Do(
		func() {
			states = []State{}
			states = append(states, newIdleState(cache, bot), newAdminSubmitState(cache, bot))
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

func (state *idleState) Handle(chatId int64, message string) error {
	mu := state.cache.AcquireLock(chatId)
	defer state.cache.ReleaseLock(chatId)
	mu.Lock()
	defer mu.Unlock()
	switch message {
	case update_handlers.ASSIGN_COMMAND:
		state.cache.Save(*interfaces.NewCachedInfo(chatId, string(ADMIN_SUBMIT_STATE)))
		return nil
	}
	return errors.Join(errors.ErrUnsupported, errors.New("answers are only to commands"))
}

func (*idleState) StateName() StateName {
	return IDLE_STATE
}

type adminSubmitState struct {
	State
	cache interfaces.HandlersCache
	bot   *tgbotapi.BotAPI
}

func newAdminSubmitState(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI) *adminSubmitState {
	return &adminSubmitState{cache: cache, bot: bot}
}

func (*adminSubmitState) StateName() StateName {
	return ADMIN_SUBMIT_STATE
}

func (state *adminSubmitState) Handle(chatId int64, message string) error {
	msg := tgbotapi.NewMessage(0, string(message))
	tgutils.SendMessageToOwners(msg, state.bot)
	return nil
}
