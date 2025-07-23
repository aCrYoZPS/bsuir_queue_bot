package sheetsapi

import (
	"fmt"
	"time"

	driveapi "github.com/aCrYoZPS/bsuir_queue_bot/src/google_docs/drive_api"
	iis_api_entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
	"google.golang.org/api/sheets/v4"
)

type SheetsApiService struct {
	groupsRepo  interfaces.GroupsRepository
	driveApi    driveapi.DriveApi
	api         *sheets.Service
}

func NewSheetsApiService(groups interfaces.GroupsRepository, driveApi driveapi.DriveApi, api *sheets.Service) *SheetsApiService {
	return &SheetsApiService{
		groupsRepo:  groups,
		driveApi:    driveApi,
		api:         api,
	}
}

type SheetsUrl = string

func (serv *SheetsApiService) CreateSheet(groupName string) (SheetsUrl, error) {
	existsRes, err := serv.driveApi.DoesSheetExist(groupName)
	if err != nil {
		return "", err
	}
	if existsRes.DoesExist() {
		sheet, err := serv.api.Spreadsheets.Get(existsRes.SpreadsheetId()).Do()
		if err != nil {
			return "", err
		}
		return sheet.SpreadsheetUrl, nil
	}
	newSheet := sheets.Spreadsheet{Properties: &sheets.SpreadsheetProperties{
		Title: groupName,
	}}

	res := serv.api.Spreadsheets.Create(&newSheet)
	sheet, err := res.Do()
	if err != nil {
		return "", err
	}

	group, err := serv.groupsRepo.GetByName(groupName)
	if err != nil {
		return "", err
	}
	group.SpreadsheetId = sheet.SpreadsheetId
	err = serv.groupsRepo.Update(group)
	if err != nil {
		return "", err
	}
	return sheet.SpreadsheetUrl, nil
}

func (serv *SheetsApiService) CreateLists(groupName string, lessons []persistance.Lesson) error {
	group, err := serv.groupsRepo.GetByName(groupName)
	if err != nil {
		return err
	}
	update := sheets.BatchUpdateSpreadsheetRequest{}
	for _, lesson := range lessons {
		updateTitle := lesson.Subject + " " + serv.formatDateToEuropean(lesson.Date)
		if iis_api_entities.Subgroup(lesson.SubgroupNumber) != iis_api_entities.AllSubgroups {
			updateTitle += fmt.Sprintf(" (%s)", fmt.Sprint(int(lesson.SubgroupNumber)))
		}
		update.Requests = append(update.Requests, &sheets.Request{
			AddSheet: &sheets.AddSheetRequest{Properties: &sheets.SheetProperties{
				Title: updateTitle,
			}},
		})
	}
	call := serv.api.Spreadsheets.BatchUpdate(group.SpreadsheetId, &update)
	_, err = call.Do()
	if err != nil {
		return err
	}

	return nil
}

func (serv *SheetsApiService) ClearSpreadsheet(spreadsheetId string) error {
	getSpreadsheetRequest := sheets.SpreadsheetsGetCall{}
	spreadsheet, err := getSpreadsheetRequest.Do()
	if err != nil {
		return err
	}
	deleteSheetsRequest := sheets.BatchUpdateSpreadsheetRequest{}
	for _, sheet := range spreadsheet.Sheets {
		deleteSheetsRequest.Requests = append(deleteSheetsRequest.Requests, &sheets.Request{
			DeleteSheet: &sheets.DeleteSheetRequest{SheetId: sheet.Properties.SheetId},
		})
	}
	call := serv.api.Spreadsheets.BatchUpdate(spreadsheetId, &deleteSheetsRequest)
	_, err = call.Do()
	return err
}

func (serv *SheetsApiService) AddLabwork(subject string, groupName string, requestedDate time.Time, sentProofTime time.Time) error {
	return nil
}


func (serv *SheetsApiService) formatDateToEuropean(date time.Time) string {
	return fmt.Sprint(date.Day()) + "." + fmt.Sprint(date.Month()) + "." + fmt.Sprint(date.Year())
}
