package sheetsapi

import (
	"context"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/labworks"
)

type SheetUrl = string
type SheetsApi interface {
	CreateSheet(ctx context.Context, groupName string, lessons []persistance.Lesson) (SheetUrl, error)
	ClearSpreadsheet(ctx context.Context, spreadsheetId string) error
	AddLabwork(context.Context, *labworks.LabworkRequest) error
}
