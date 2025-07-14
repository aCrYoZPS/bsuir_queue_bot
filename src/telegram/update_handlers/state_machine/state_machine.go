package stateMachine

import (
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers"
)

type StateName string

type State interface {
	StateName() StateName
	Handle(cache interfaces.HandlersCache, newState StateName) error
}

type StateMachine struct {
	update_handlers.StateMachine
	cache interfaces.HandlersCache
}

func NewStateMachine(cache interfaces.HandlersCache) *StateMachine {
	return &StateMachine{cache: cache}
}

func (machine *StateMachine) HandleState(chatId int64, message string) error {
	info, err := machine.cache.Get(chatId)
	if err != nil {
		return err
	}
	state, err := getStateByName(info.State())
	if err != nil {
		return err
	}
	state.Handle(machine.cache, StateName(message))
	return nil
}
