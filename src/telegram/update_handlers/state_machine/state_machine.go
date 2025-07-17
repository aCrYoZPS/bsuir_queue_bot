package stateMachine

import (
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type StateName string

type State interface {
	StateName() StateName
	Handle(chatId int64, message *tgbotapi.Message) error
}

type StateMachine struct {
	update_handlers.StateMachine
	cache interfaces.HandlersCache
	bot   *tgbotapi.BotAPI
}

func NewStateMachine(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI) *StateMachine {
	InitStates(cache, bot)
	return &StateMachine{cache: cache, bot: bot}
}

func (machine *StateMachine) HandleState(chatId int64, message *tgbotapi.Message) error {
	mu := machine.cache.AcquireLock(chatId)
	mu.Lock()

	defer mu.Unlock()
	defer machine.cache.ReleaseLock(chatId)

	info, err := machine.cache.GetState(chatId)
	if err != nil {
		return err
	}

	var state State
	if info != nil {
		state, err = getStateByName(info.State())
		if err != nil {
			return err
		}
	} else {
		state, err = getStateByName(string(IDLE_STATE))
		if err != nil {
			return err
		}
	}

	return state.Handle(chatId, message)
}
