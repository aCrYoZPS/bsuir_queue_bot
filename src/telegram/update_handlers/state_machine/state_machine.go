package stateMachine

import (
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers"
)

type StateName string

type State interface {
	Handle(machine update_handlers.StateMachine, newState StateName) error
}

type StateMachine struct {
	state State
	update_handlers.StateMachine
	cache interfaces.CachedInfo
}

func NewStateMachine(cache interfaces.CachedInfo) *StateMachine {
	return &StateMachine{cache: cache, state: &DefaultState{}}
}

func (machine *StateMachine) HandleState(curState, message string) error {
	return nil
}
