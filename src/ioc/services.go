package ioc

import (
	"context"
	"log/slog"

	google_docs_auth "github.com/aCrYoZPS/bsuir_queue_bot/src/google_docs/auth"
	sheetsapi "github.com/aCrYoZPS/bsuir_queue_bot/src/google_docs/sheets_api"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var useSheetsApi = provider(
	func() *sheets.Service {
		client, err := google_docs_auth.GetClient()
		if err != nil {
			slog.Error(err.Error())
			panic(err)
		}
		ctx := context.Background()
		srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			slog.Error(err.Error())
			panic(err)
		}
		return srv
	},
)

var UseSheetsApiService = provider(
	func() sheetsapi.SheetsApi {
		return sheetsapi.NewSheetsApiService(
			useGroupsGepository(), useLessonsRepository(),
			useSheetsApi(),
		)
	},
)
