package ioc

import (
	"context"
	"log/slog"
	"net/http"

	google_docs_auth "github.com/aCrYoZPS/bsuir_queue_bot/src/google_docs/auth"
	driveapi "github.com/aCrYoZPS/bsuir_queue_bot/src/google_docs/drive_api"
	sheetsapi "github.com/aCrYoZPS/bsuir_queue_bot/src/google_docs/sheets_api"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/ioc/constants"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/logging"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/bot"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers"
	stateMachine "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var useGroupsService = provider(
	func() *iis_api.GroupsService {
		ctx, cancel := context.WithTimeout(context.Background(), constants.INIT_TIMEOUT)
		defer cancel()
		serv := iis_api.NewGroupsService(useGroupsRepository())
		err := serv.InitAllGroups(ctx)
		if err != nil {
			logging.FatalLog(err.Error())
		}
		return serv
	},
)

var useGoogleClient = provider(
	func() *http.Client {
		client, err := google_docs_auth.GetClient()
		if err != nil {
			logging.FatalLog(err.Error())
		}
		return client
	},
)

var useSheetsApi = provider(
	func() *sheets.Service {
		ctx := context.Background()
		srv, err := sheets.NewService(ctx, option.WithHTTPClient(useGoogleClient()))
		if err != nil {
			slog.Error(err.Error())
			panic(err)
		}
		return srv
	},
)

var useDriveApi = provider(
	func() *drive.Service {
		ctx := context.Background()
		srv, err := drive.NewService(ctx, option.WithHTTPClient(useGoogleClient()))
		if err != nil {
			logging.FatalLog(err.Error())
		}
		return srv
	},
)

var UseDriveApiService = provider(
	func() *driveapi.DriveApiService {
		return driveapi.NewDriveApiService(
			useGroupsRepository(), useDriveApi(),
		)
	},
)

var UseSheetsApiService = provider(
	func() *sheetsapi.SheetsApiService {
		return sheetsapi.NewSheetsApiService(
			useGroupsRepository(),
			UseDriveApiService(), useSheetsApi(),
		)
	},
)

var UseMessageService = provider(
	func() bot.MessagesService {
		return update_handlers.NewMessagesHandler(
			useStateMachine(), useHandlersCache(),
		)
	},
)

var UseLessonsService = provider(
	func() *iis_api.LessonsService {
		return iis_api.NewLessonsService(useLessonsRepository(), UseSheetsApiService())
	},
)

var UseCallbacksService = provider(
	func() bot.CallbacksService {
		return stateMachine.NewCallbackService(
			useUsersRepository(), useHandlersCache(), useLessonsRequestsRepository(),
			useRequestsRepository(), useAdminRequestsRepository(), UseLessonsService(), UseSheetsApiService(), useLessonsRepository())
	},
)
