package stateMachine

import (
	"errors"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers"
)

type DefaultState struct {
	State
}

func (*DefaultState) Handle(machine update_handlers.StateMachine, message string) error {
	return errors.Join(errors.ErrUnsupported, errors.New("answers are only to commands"))
}

type AdminInfoState struct {
	State
}

func Handle(machine update_handlers.StateMachine, message string) {}
