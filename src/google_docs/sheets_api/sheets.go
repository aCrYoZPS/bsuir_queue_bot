package sheetsapi

import (
	"fmt"
	"time"

	iis_api_entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"google.golang.org/api/sheets/v4"
)

type SheetsApiService struct {
	groupsRepo  interfaces.GroupsRepository
	lessonsRepo interfaces.LessonsRepository
	api         *sheets.Service
}

func NewSheetsApiService(groups interfaces.GroupsRepository, lessons interfaces.LessonsRepository, api *sheets.Service) *SheetsApiService {
	return &SheetsApiService{
		groupsRepo:  groups,
		lessonsRepo: lessons,
		api:         api,
	}
}

func (serv *SheetsApiService) CreateSheets() error {
	groups, err := serv.groupsRepo.GetAll()
	if err != nil {
		return err
	}

	for _, group := range groups {
		newSheet := sheets.Spreadsheet{Properties: &sheets.SpreadsheetProperties{
			Title: group.Name,
		}}

		res := serv.api.Spreadsheets.Create(&newSheet)
		_, err := res.Do()
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
		var update = sheets.BatchUpdateSpreadsheetRequest{}

		lessons, err := serv.lessonsRepo.GetAllLabworks(int64(group.Id))
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
	var getSpreadsheetRequest = sheets.SpreadsheetsGetCall{}
	spreadsheet, err := getSpreadsheetRequest.Do()
	if err != nil {
		return err
	}
	var deleteSheetsRequest = sheets.BatchUpdateSpreadsheetRequest{}
	for _, sheet := range spreadsheet.Sheets {
		deleteSheetsRequest.Requests = append(deleteSheetsRequest.Requests, &sheets.Request{
			DeleteSheet: &sheets.DeleteSheetRequest{SheetId: sheet.Properties.SheetId},
		})
	}
	call := serv.api.Spreadsheets.BatchUpdate(spreadsheetId, &deleteSheetsRequest)
	_, err = call.Do()
	return err
}

// 24.04.2005 format of date
func (serv *SheetsApiService) formatDateToEuropean(date time.Time) string {
	return fmt.Sprint(date.Day()) + "." + fmt.Sprint(date.Month()) + "." + fmt.Sprint(date.Year())
}

func (serv *SheetsApiService) ClearLists() error {
	return nil
}
