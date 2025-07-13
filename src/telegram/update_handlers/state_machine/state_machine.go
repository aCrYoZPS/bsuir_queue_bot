package stateMachine

import (
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers"
)

type StateName string

type State interface {
	Handle(machine StateMachine, newState StateName)
}

type StateMachine struct {
	state State
	update_handlers.StateMachine
	cache interfaces.CachedInfo
}

func NewStateMachine(cache interfaces.CachedInfo) *StateMachine {
	return &StateMachine{cache: cache}
}

func HandleState(curState, message string) error {
	return nil
}
