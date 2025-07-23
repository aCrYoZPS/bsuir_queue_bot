package sheetsapi

import (
	"time"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
)

type SheetUrl = string
type SheetsApi interface {
	CreateSheet(groupName string) (SheetUrl, error)
	ClearSpreadsheet(spreadsheetId string) error
	CreateLists(groupName string, lessons []persistance.Lesson) error
	AddLabwork(subject string, groupName string, requestedDate time.Time, sentProofTime time.Time) error
}
