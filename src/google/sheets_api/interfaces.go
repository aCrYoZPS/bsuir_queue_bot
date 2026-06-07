package sheetsapi

import (
	"context"
	"time"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistence"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/labworks"
)

type SheetUrl = string
type SheetsApi interface {
	Add(ctx context.Context, lesson *persistence.Lesson) error
	CreateSheet(ctx context.Context, groupName string, lessons []persistence.Lesson) (SheetUrl, error)
	ClearSpreadsheet(ctx context.Context, spreadsheetId string, before time.Time) error
	AddLabworkRequest(context.Context, *labworks.AppendedLabwork) error
}
