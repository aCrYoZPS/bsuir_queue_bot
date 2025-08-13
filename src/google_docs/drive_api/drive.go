package driveapi

import (
	"context"
	"fmt"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"google.golang.org/api/drive/v3"
)

var _ DriveApi = (*DriveApiService)(nil)

type DriveApiService struct {
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

func (serv *DriveApiService) GetSpreadsheets(ctx context.Context) ([]string, error) {
	files, err := serv.api.Files.List().Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	spreadsheetIds := []string{}
	nextPage := true
	for nextPage {
		for _, file := range files.Files {
			if file.MimeType == SHEETS_MIME_TYPE {
				spreadsheetIds = append(spreadsheetIds, file.Id)
			}
		}
		if files.NextPageToken == "" {
			nextPage = false
		} else {
			files, err = serv.api.Files.List().Context(ctx).PageToken(files.NextPageToken).Do()
			if err != nil {
				return nil, err
			}
		}
	}
	return spreadsheetIds, nil
}

func (serv *DriveApiService) SetSpreadsheetPermissions(ctx context.Context, spreadsheetId string) error {
	_, err := serv.api.Permissions.Create(spreadsheetId, &drive.Permission{Type: "anyone", Role: "reader"}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to set spreadsheet permissions: %w", err)
	}
	return nil
}
