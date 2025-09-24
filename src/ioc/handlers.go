package ioc

import (
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers"
	stateMachine "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine"
)

var useStateMachine = provider(
	func() update_handlers.StateMachine {
		stateMachine := stateMachine.NewStateMachine(
			stateMachine.NewStatesConfig(nil,
				useHandlersCache(), useTgBot(),
				useGroupsService(), useUsersRepository(),
				useRequestsRepository(), useAdminRequestsRepository(),
				useLessonsService()),
		)
		stateMachine.StateMachine = stateMachine
		return stateMachine
	},
)
