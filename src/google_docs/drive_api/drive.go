package driveapi

import (
	"context"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"google.golang.org/api/drive/v3"
)

var _ DriveApi = (*DriveApiService)(nil)
type DriveApiService struct {
	DriveApi
	groupsRepo interfaces.GroupsRepository
	api        *drive.Service
}

func NewDriveApiService(groups interfaces.GroupsRepository, api *drive.Service) *DriveApiService {
	return &DriveApiService{
		groupsRepo: groups,
		api:        api,
	}
}

const SHEETS_MIME_TYPE = "application/vnd.google-apps.spreadsheet"

func (serv *DriveApiService) DoesSheetExist(ctx context.Context, name string) (SpreadsheetResult, error) {
	files, err := serv.api.Files.List().Context(ctx).Do()
	if err != nil {
		return SpreadsheetResult{}, err
	}
	nextPage := true
	for nextPage {
		for _, file := range files.Files {
			if file.MimeType == SHEETS_MIME_TYPE && file.Name == name {
				return SpreadsheetResult{doesExist: true, spreadsheetId: file.Id}, nil
			}
		}
		if files.NextPageToken == "" {
			nextPage = false
		} else {
			files, err = serv.api.Files.List().PageToken(files.NextPageToken).Do()
			if err != nil {
				return SpreadsheetResult{}, err
			}
		}
	}
	return SpreadsheetResult{}, nil
}
