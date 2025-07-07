package sheetsapi

import (
	"fmt"
	"time"

	driveapi "github.com/aCrYoZPS/bsuir_queue_bot/src/google_docs/drive_api"
	iis_api_entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"google.golang.org/api/sheets/v4"
)

type SheetsApiService struct {
	groupsRepo  interfaces.GroupsRepository
	lessonsRepo interfaces.LessonsRepository
	driveApi    driveapi.DriveApi
	api         *sheets.Service
}

func NewSheetsApiService(groups interfaces.GroupsRepository, lessons interfaces.LessonsRepository, driveApi driveapi.DriveApi, api *sheets.Service) *SheetsApiService {
	return &SheetsApiService{
		groupsRepo:  groups,
		lessonsRepo: lessons,
		driveApi:    driveApi,
		api:         api,
	}
}

func (serv *SheetsApiService) CreateSheets() error {
	groups, err := serv.groupsRepo.GetAll()
	if err != nil {
		return err
	}

	// I haven't figured out a way to batch these requests :(
	for _, group := range groups {
		exists, err := serv.driveApi.DoesSheetExist(group.Name)
		if err != nil {
			return nil
		}
		if exists {
			continue
		}
		newSheet := sheets.Spreadsheet{Properties: &sheets.SpreadsheetProperties{
			Title: group.Name,
		}}

		res := serv.api.Spreadsheets.Create(&newSheet)
		sheet, err := res.Do()
		if err != nil {
			return err
		}
		group.SpreadsheetId = sheet.SpreadsheetId
		err = serv.groupsRepo.Update(&group)
		if err != nil {
			return err
		}
	}
	return nil
}

func (serv *SheetsApiService) CreateLists() error {
	groups, err := serv.groupsRepo.GetAll()
	if err != nil {
		return err
	}
	for _, group := range groups {
		update := sheets.BatchUpdateSpreadsheetRequest{}
		lessons, err := serv.lessonsRepo.GetAll(int64(group.Id))
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

func (serv *SheetsApiService) formatDateToEuropean(date time.Time) string {
	return fmt.Sprint(date.Day()) + "." + fmt.Sprint(date.Month()) + "." + fmt.Sprint(date.Year())
}
