package sheetsapi

import (
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/labworks"
)

type SheetUrl = string
type SheetsApi interface {
	CreateSheet(groupName string) (SheetUrl, error)
	ClearSpreadsheet(spreadsheetId string) error
	CreateLists(groupName string, lessons []persistance.Lesson) error
	AddLabwork(*labworks.LabworkRequest) error
}
