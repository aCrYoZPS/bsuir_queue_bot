package stateMachine

import (
	"errors"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
)

const (
	SELECT_SUBJECT_STATE StateName = "subject"
	SELECT_DATE_STATE    StateName = "date"
	SUBMIT_PROOF_STATE   StateName = "proof"
	ADMIN_SUBMIT_STATE   StateName = "submit"
	IDLE_STATE           StateName = ""
)

var states = []State{&idleState{}, &adminSubmitState{}}

func getStateByName(name string) (State, error) {
	for _, state := range states {
		if state.StateName() == StateName(name) {
			return state, nil
		}
	}
	return &idleState{}, errors.New("no such state")
}

type idleState struct {
	State
}

func (*idleState) Handle(machine interfaces.HandlersCache, message StateName) error {
	return errors.Join(errors.ErrUnsupported, errors.New("answers are only to commands"))
}

func (*idleState) StateName() StateName {
	return IDLE_STATE
}

type adminSubmitState struct {
	State
}

func (*adminSubmitState) StateName() StateName {
	return ADMIN_SUBMIT_STATE
}

func (*adminSubmitState) Handle(cache interfaces.HandlersCache, message StateName) error {
	return nil
}
