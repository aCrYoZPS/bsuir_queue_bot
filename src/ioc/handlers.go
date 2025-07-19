package ioc

import (
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers"
	stateMachine "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine"
)

var useStateMachine = provider(
	func() update_handlers.StateMachine {
		return stateMachine.NewStateMachine(
			stateMachine.NewStatesConfig(useHandlersCache(),
				useTgBot(), useGroupsService()),
		)
	},
)
