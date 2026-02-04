package ioc

import (
	"context"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/cron"
	stateMachine "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers"
	delete "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/admin/delete_user"
	admin "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/admin_submit"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/constants"
	customlabworks "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/custom_labworks"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/group"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/labworks"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/queue"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func RegisterRoutes(mux *tgutils.Mux) {
	mux.NotFoundHandler = useIdleState()

	adminMux := tgutils.NewMux(useHandlersCache(), useTgBot())
	mux.RegisterRoute(constants.ADMIN_STATES, useAdminMiddleware(adminMux)())
	RegisterDeleteRoutes(adminMux)

	mux.RegisterRoute(constants.IDLE_STATE, (useIdleState()))
	RegisterAdminSubmitRoutes(mux)
	RegisterLabworkRoutes(mux)
	RegisterLabworkAddRoutes(mux)
	RegisterGroupRoutes(mux)
	RegisterQueueRoutes(mux)
	RegisterCronCalbacks(mux)
}

func RegisterLabworkRoutes(mux *tgutils.Mux) {
	mux.RegisterRoute(constants.LABWORK_SUBMIT_START_STATE, useLabworkSubmitStartState())
	mux.RegisterRoute(constants.LABWORK_SUBMIT_NUMBER_STATE, useLabworkSubmitNumberState())
	mux.RegisterRoute(constants.LABWORK_SUBMIT_PROOF_STATE, useLabworkSubmitProofState())
	mux.RegisterRoute(constants.LABWORK_SUBMIT_WAITING_STATE, useLabworkSubmitWaitingState())

	mux.RegisterCallback(constants.LABWORK_CALLBACKS, useLabworkSubmitCallbackHandler())
}

func RegisterLabworkAddRoutes(mux *tgutils.Mux) {
	mux.RegisterRoute(constants.LABWORK_ADD_START_STATE, useLabworkAddStartState())
	mux.RegisterRoute(constants.LABWORK_ADD_SUBMIT_NAME_STATE, useLabworkAddSubmitNameState())
	mux.RegisterRoute(constants.LABWORK_ADD_WAITING_STATE, useLabworkAddWaitingState())

	mux.RegisterCallback(constants.CALENDAR_CALLBACKS, useCalendarCallbackHandler())
	mux.RegisterCallback(constants.TIME_PICKER_CALLBACKS, useTimePickerCallbackHandler())
	mux.RegisterCallback(constants.IGNORE_CALLBACKS, tgutils.CallbackHandlerFunc(func(ctx context.Context, update *tgbotapi.Update, bot *tgutils.Bot) error { return nil }))
}

func RegisterGroupRoutes(mux *tgutils.Mux) {
	mux.RegisterRoute(constants.GROUP_SUBMIT_START_STATE, useGroupSubmitStartState())
	mux.RegisterRoute(constants.GROUP_WAITING_STATE, useGroupSubmitWaitingState())
	mux.RegisterRoute(constants.GROUP_SUBMIT_NAME_STATE, useGroupSubmitNameState())
	mux.RegisterRoute(constants.GROUP_SUBMIT_GROUPNAME_STATE, useGroupSubmitGroupNameState())

	mux.RegisterCallback(constants.GROUP_CALLBACKS, useGroupCallbackHandler())
}

func RegisterQueueRoutes(mux *tgutils.Mux) {
	mux.RegisterRoute(constants.QUEUE_START_STATE, useQueueStartState())
	mux.RegisterRoute(constants.QUEUE_WAITING_STATE, useQueueWaitingState())
	mux.RegisterCallback(constants.QUEUE_CALLBACKS, useQueueCallbackHandler())
}

func RegisterAdminSubmitRoutes(mux *tgutils.Mux) {
	mux.RegisterRoute(constants.ADMIN_SUBMIT_START_STATE, useAdminSubmitStartState())
	mux.RegisterRoute(constants.ADMIN_SUBMITTING_NAME_STATE, useAdminSubmittingNameState())
	mux.RegisterRoute(constants.ADMIN_SUBMITTING_GROUP_STATE, useAdminSubmittingGroupState())
	mux.RegisterRoute(constants.ADMIN_SUBMITTING_PROOF_STATE, useAdminSubmittingProofState())
	mux.RegisterRoute(constants.ADMIN_WAITING_STATE, useAdminWaitingState())

	mux.RegisterCallback(constants.ADMIN_CALLBACKS, useAdminCallbackHandler())
}

func RegisterDeleteRoutes(mux *tgutils.Mux) {
	mux.RegisterRoute(constants.DELETE_START_STATE, useDeleteStartState())
	mux.RegisterRoute(constants.DELETE_CHOOSE_STATE, useDeleteChooseState())
}

func RegisterCronCalbacks(mux *tgutils.Mux) {
	mux.RegisterCallback(cron.REMINDER_CALLBACKS, useReminderCallbackHandler())
}

var useLabworkSubmitStartState = provider(
	func() tgutils.MuxHandler {
		return labworks.NewLabworkSubmitStartState(useTgBot(), useHandlersCache(), useLessonsRepository(), useUsersRepository())
	},
)
var useLabworkSubmitNumberState = provider(
	func() tgutils.MuxHandler {
		return labworks.NewLabworkSubmitNumberState(useTgBot(), useHandlersCache(), useLessonsRepository(), useUsersRepository(), useMux())
	},
)
var useLabworkSubmitProofState = provider(
	func() tgutils.MuxHandler {
		return labworks.NewLabworkSubmitProofState(useTgBot(), useHandlersCache(), useGroupsService(), useRequestsRepository(), useLessonsRequestsRepository())
	},
)
var useLabworkSubmitWaitingState = provider(
	func() tgutils.MuxHandler {
		return labworks.NewLabworkSubmitWaitingState(useTgBot(), useHandlersCache(), useMux())
	},
)
var useLabworkSubmitCallbackHandler = provider(
	func() tgutils.CallbackHandler {
		return labworks.NewLabworksCallbackHandler(useTgBot(), useHandlersCache(), useLessonsRepository(), useRequestsRepository(),
			useLessonsRequestsRepository(), useUsersRepository(), UseSheetsApiService())
	},
)
var useLabworkAddStartState = provider(
	func() tgutils.MuxHandler {
		return customlabworks.NewLabworkAddStartState(useTgBot(), useHandlersCache(), useUsersRepository())
	},
)
var useLabworkAddSubmitNameState = provider(
	func() tgutils.MuxHandler {
		return customlabworks.NewLabworkAddSubmitNameState(useTgBot(), useHandlersCache())
	},
)
var useLabworkAddWaitingState = provider(
	func() tgutils.MuxHandler {
		return customlabworks.NewLabworkAddWaitingState(useTgBot(), useHandlersCache(), useMux())
	},
)
var useCalendarCallbackHandler = provider(
	func() tgutils.CallbackHandler {
		return customlabworks.NewCalendarCallbackHandler(useTgBot(), useHandlersCache())
	},
)
var useTimePickerCallbackHandler = provider(
	func() tgutils.CallbackHandler {
		return customlabworks.NewTimePickerCallbackHandler(useTgBot(), useLessonsRepository(), useHandlersCache())
	},
)

var useGroupSubmitStartState = provider(
	func() tgutils.MuxHandler {
		return group.NewGroupSubmitState(useHandlersCache(), useTgBot(), useGroupsRepository(), useUsersRepository())
	},
)
var useGroupSubmitNameState = provider(
	func() tgutils.MuxHandler {
		return group.NewGroupSubmitNameState(useHandlersCache(), useTgBot(), useGroupsService(), useRequestsRepository(), useMux())
	},
)
var useGroupSubmitGroupNameState = provider(
	func() tgutils.MuxHandler {
		return group.NewGroupSubmitGroupNameState(useHandlersCache(), useTgBot(), useGroupsService())
	},
)
var useGroupSubmitWaitingState = provider(
	func() tgutils.MuxHandler {
		return group.NewGroupWaitingState(useHandlersCache(), useTgBot())
	},
)
var useGroupCallbackHandler = provider(
	func() tgutils.CallbackHandler {
		return group.NewGroupCallbackHandler(useUsersRepository(), useHandlersCache(), useRequestsRepository())
	},
)

var useQueueStartState = provider(
	func() tgutils.MuxHandler {
		return queue.NewQueueStartState(useTgBot(), useHandlersCache(), useUsersRepository(), useLessonsRepository())
	},
)
var useQueueWaitingState = provider(
	func() tgutils.MuxHandler {
		return queue.NewQueueWaitingState(useHandlersCache(), useTgBot())
	},
)
var useQueueCallbackHandler = provider(
	func() tgutils.CallbackHandler {
		return queue.NewQueueCallbackHandler(useUsersRepository(), useLessonsRepository(), useHandlersCache(), useTgBot(), useLessonsRequestsRepository())
	},
)
var useAdminSubmitStartState = provider(
	func() tgutils.MuxHandler {
		return admin.NewAdminSubmitState(useHandlersCache(), useTgBot(), useUsersRepository())
	},
)
var useAdminSubmittingNameState = provider(
	func() tgutils.MuxHandler {
		return admin.NewAdminSubmittingNameState(useHandlersCache(), useTgBot(), useMux())
	},
)
var useAdminSubmittingGroupState = provider(
	func() tgutils.MuxHandler {
		return admin.NewAdminSubmitingGroupState(useHandlersCache(), useTgBot(), useGroupsRepository(), useMux())
	},
)
var useAdminSubmittingProofState = provider(
	func() tgutils.MuxHandler {
		return admin.NewAdminSubmitingProofState(useHandlersCache(), useTgBot(), useAdminRequestsRepository(), useMux())
	},
)
var useAdminWaitingState = provider(
	func() tgutils.MuxHandler {
		return admin.NewAdminWaitingProofState(useHandlersCache(), useTgBot())
	},
)
var useAdminCallbackHandler = provider(
	func() tgutils.CallbackHandler {
		return admin.NewAdminCallbackHandler(useUsersRepository(), useHandlersCache(), useAdminRequestsRepository(), UseLessonsService())
	},
)

var useIdleState = provider(
	func() tgutils.MuxHandler {
		return stateMachine.NewIdleState(useHandlersCache(), useTgBot(), useUsersRepository(), useGroupsRepository(), useLessonsRepository(), useMux())
	},
)

var useDeleteStartState = provider(
	func() *delete.DeleteStartState {
		return delete.NewDeleteStartState(useTgBot(), useUsersRepository(), useHandlersCache())
	},
)
var useDeleteChooseState = provider(
	func() *delete.DeleteChooseState {
		return delete.NewDeleteChooseState(useTgBot(), useHandlersCache(), useUsersRepository())
	},
)

var useReminderCallbackHandler = provider(func() *cron.ReminderCallbackHandler {
	return cron.NewSheetsRefreshCallbackHandler(useLessonsRequestsRepository(), UseSheetsApiService(), useUsersRepository(), UseLessonsService())
})
