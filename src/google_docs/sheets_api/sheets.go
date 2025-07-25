package sheetsapi

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	driveapi "github.com/aCrYoZPS/bsuir_queue_bot/src/google_docs/drive_api"
	iis_api_entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/labworks"
	"google.golang.org/api/sheets/v4"
)

var _ SheetsApi = (*SheetsApiService)(nil)
var _ labworks.SheetsService = (*SheetsApiService)(nil)

type SheetsApiService struct {
	groupsRepo interfaces.GroupsRepository
	driveApi   driveapi.DriveApi
	api        *sheets.Service
}

func NewSheetsApiService(groups interfaces.GroupsRepository, driveApi driveapi.DriveApi, api *sheets.Service) *SheetsApiService {
	return &SheetsApiService{
		groupsRepo: groups,
		driveApi:   driveApi,
		api:        api,
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
	newSpreadsheet := sheets.Spreadsheet{Properties: &sheets.SpreadsheetProperties{
		Title: groupName,
	}}

	res := serv.api.Spreadsheets.Create(&newSpreadsheet)
	spreadsheet, err := res.Do()
	if err != nil {
		return "", err
	}

	group, err := serv.groupsRepo.GetByName(groupName)
	if err != nil {
		return "", err
	}
	group.SpreadsheetId = spreadsheet.SpreadsheetId
	err = serv.groupsRepo.Update(group)
	if err != nil {
		return "", err
	}
	return spreadsheet.SpreadsheetUrl, nil
}

func (serv *SheetsApiService) CreateLists(groupName string, lessons []persistance.Lesson) error {
	group, err := serv.groupsRepo.GetByName(groupName)
	if err != nil {
		return err
	}
	update := sheets.BatchUpdateSpreadsheetRequest{}
	for _, lesson := range lessons {
		updateTitle := serv.createLessonName(lesson)
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

func (serv *SheetsApiService) createLessonName(lesson persistance.Lesson) string {
	return lesson.Subject + " " + serv.formatDateToEuropean(lesson.Date)
}

func (serv *SheetsApiService) formatDateToEuropean(date time.Time) string {
	zeroPrependedDay := fmt.Sprint(date.Day()/10) + fmt.Sprint(date.Day()%10)
	zeroPrependedMonth := fmt.Sprint(int(date.Month())/10) + fmt.Sprint(int(date.Month())%10)
	return zeroPrependedDay + "." + zeroPrependedMonth + "." + fmt.Sprint(date.Year())
}

func parseLessonName(name string) (subject string, date time.Time) {
	data := strings.Split(name, " ")
	if len(data) != 2 {
		return "", time.Time{}
	}
	subject = data[0]
	datePoints := strings.Split(data[2], ".")
	if len(datePoints) != 3 {
		return "", time.Time{}
	}
	day, err := strconv.Atoi(datePoints[0])
	if err != nil {
		return "", time.Time{}
	}
	month, err := strconv.Atoi(datePoints[1])
	if err != nil {
		return "", time.Time{}
	}
	year, err := strconv.Atoi(datePoints[2])
	if err != nil {
		return "", time.Time{}
	}
	date = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return subject, date
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

func (serv *SheetsApiService) AddLabwork(req *labworks.LabworkRequest) error {
	group, err := serv.groupsRepo.GetByName(req.GroupName)
	if err != nil {
		return err
	}
	spreadsheetId := group.SpreadsheetId
	spreadsheet, err := serv.api.Spreadsheets.Get(spreadsheetId).Do()
	if err != nil {
		return err
	}
	for _, sheet := range spreadsheet.Sheets {
		titleSubject, titleDate := parseLessonName(sheet.Properties.Title)
		if titleSubject == req.DisciplineName && req.RequestedDate.Equal(titleDate) {
			if sheet.Tables == nil {
				err = serv.createTable(sheet)
				if err != nil {
					return err
				}
			}
			err = serv.appendToSheet(spreadsheetId, sheet, req)
			return err
		}
	}
	return errors.New("no such labwork found")
}

func (serv *SheetsApiService) createTable(sheet *sheets.Sheet) error {
	_, err := serv.api.Spreadsheets.BatchUpdate("", &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{{
			UpdateTable: &sheets.UpdateTableRequest{
				Fields: "*",
				Table: &sheets.Table{
					Name: "Очередь",
					Range: &sheets.GridRange{
						SheetId:          sheet.Properties.SheetId,
						StartRowIndex:    0,
						EndRowIndex:      1,
						StartColumnIndex: 0,
						EndColumnIndex:   3,
					},
					ColumnProperties: []*sheets.TableColumnProperties{
						{
							ColumnIndex: 0,
							ColumnName:  "Фамилия и имя",
							ColumnType:  "TEXT",
						},
						{
							ColumnIndex: 1,
							ColumnName:  "Номер лабораторной",
							ColumnType:  "TEXT",
						},
						{
							ColumnIndex: 2,
							ColumnName:  "Дата и время заявки",
							ColumnType:  "DATE_TIME",
						},
					},
				},
			},
		},
		}},
	).Do()
	return err
}

func (serv *SheetsApiService) appendToSheet(spreadsheetId string, sheet *sheets.Sheet, req *labworks.LabworkRequest) error {
	tableSearchRange := fmt.Sprintf("'%s'!A1:B5", sheet.Properties.Title)
	_, err := serv.api.Spreadsheets.Values.Append(spreadsheetId, tableSearchRange, &sheets.ValueRange{
		Range:          tableSearchRange,
		MajorDimension: "ROWS",
		Values:         [][]any{{req.FullName, req.LabworkNumber, serv.formatDateTimeToEuropean(req.SentProofTime)}},
	}).Do()
	return err
}

func (serv *SheetsApiService) formatDateTimeToEuropean(dateTime time.Time) string {
	zeroPrependedHour := fmt.Sprint(dateTime.Hour()/10) + fmt.Sprint(dateTime.Hour()%10)
	zeroPrependedMinutes := fmt.Sprint(dateTime.Minute()/10) + fmt.Sprint(dateTime.Minute()%10)
	zeroPrependedSeconds := fmt.Sprint(dateTime.Second()/10) + fmt.Sprint(dateTime.Second()%10)

	date := fmt.Sprint(dateTime.Day()) + "." + fmt.Sprint(dateTime.Month()) + "." + fmt.Sprint(dateTime.Year())
	time := zeroPrependedHour + ":" + zeroPrependedMinutes + ":" + zeroPrependedSeconds
	return date + " " + time
}
