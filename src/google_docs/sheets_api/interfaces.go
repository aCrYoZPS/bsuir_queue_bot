package sheetsapi

type SheetsApi interface {
	CreateSheets() error
	CreateLists() error
	ClearSpreadsheet(spreadsheetId string) error
}
