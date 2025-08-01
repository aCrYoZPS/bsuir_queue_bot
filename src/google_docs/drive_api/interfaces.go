package driveapi

import "context"

type SpreadsheetResult struct {
	doesExist     bool
	spreadsheetId string
}

func (res *SpreadsheetResult) DoesExist() bool {
	return res.doesExist
}

func (res *SpreadsheetResult) SpreadsheetId() string {
	return res.spreadsheetId
}

type DriveApi interface {
	DoesSheetExist(ctx context.Context, name string) (SpreadsheetResult, error)
}
