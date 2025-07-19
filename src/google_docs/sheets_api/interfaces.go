package sheetsapi

type SheetUrl = string
type SheetsApi interface {
	CreateSheet(groupName string) (SheetUrl, error)
	ClearSpreadsheet(spreadsheetId string) error
}
