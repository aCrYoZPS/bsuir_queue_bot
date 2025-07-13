package ioc

import (
	"context"
	"log/slog"
	"net/http"

	google_docs_auth "github.com/aCrYoZPS/bsuir_queue_bot/src/google_docs/auth"
	driveapi "github.com/aCrYoZPS/bsuir_queue_bot/src/google_docs/drive_api"
	sheetsapi "github.com/aCrYoZPS/bsuir_queue_bot/src/google_docs/sheets_api"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/logging"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/bot"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
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
	func() driveapi.DriveApi {
		return driveapi.NewDriveApiService(
			useGroupsGepository(), useDriveApi(),
		)
	},
)

var UseSheetsApiService = provider(
	func() sheetsapi.SheetsApi {
		return sheetsapi.NewSheetsApiService(
			useMockGroupsRepository(), useLessonsRepository(),
			UseDriveApiService(), useSheetsApi(),
		)
	},
)

var UseMessageService = provider(
	func() bot.MessagesService {
		return update_handlers.NewMessagesHandler(
			useHandlersCache(),
		)
	},
)

var UseCallbacksService = provider(
	func() bot.CallbacksService {
		return update_handlers.NewCallbackService()
	},
)
