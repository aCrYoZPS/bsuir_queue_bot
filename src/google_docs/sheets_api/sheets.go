package sheetsapi

import (
	"context"
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

var errSheetExists = errors.New("sheets: sheet with such name already exists")

func ErrSheetsExists() error {
	return errSheetExists
}

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

func (serv *SheetsApiService) CreateSheet(ctx context.Context, groupName string, lessons []persistance.Lesson) (SheetsUrl, error) {
	group, err := serv.groupsRepo.GetByName(ctx, groupName)
	if err != nil {
		return "", err
	}

	existsRes, err := serv.driveApi.DoesSheetExist(ctx, groupName)
	if err != nil {
		return "", err
	}
	if existsRes.DoesExist() {
		if group.SpreadsheetId == "" {
			group.SpreadsheetId = existsRes.SpreadsheetId()
		}
		err := serv.groupsRepo.Update(ctx, group)
		if err != nil {
			return "", err
		}
		sheet, err := serv.api.Spreadsheets.Get(existsRes.SpreadsheetId()).Context(ctx).Do()
		if err != nil {
			return "", err
		}
		return sheet.SpreadsheetUrl, nil
	}
	newSpreadsheet := sheets.Spreadsheet{Properties: &sheets.SpreadsheetProperties{
		Title: groupName,
		Locale: "ru",
	}}

	res := serv.api.Spreadsheets.Create(&newSpreadsheet)
	spreadsheet, err := res.Context(ctx).Do()
	if err != nil {
		return "", err
	}

	group.SpreadsheetId = spreadsheet.SpreadsheetId
	err = serv.groupsRepo.Update(ctx, group)
	if err != nil {
		return "", err
	}

	err = serv.createLists(ctx, groupName, lessons)
	if err != nil {
		return "", err
	}
	return spreadsheet.SpreadsheetUrl, nil
}

func (serv *SheetsApiService) createLists(ctx context.Context, groupName string, lessons []persistance.Lesson) error {
	group, err := serv.groupsRepo.GetByName(ctx, groupName)
	if err != nil {
		return err
	}
	update := sheets.BatchUpdateSpreadsheetRequest{}
	for _, lesson := range lessons {
		updateTitle := serv.createLessonName(lesson)
		update.Requests = append(update.Requests, &sheets.Request{
			AddSheet: &sheets.AddSheetRequest{Properties: &sheets.SheetProperties{
				Title: updateTitle,
			}},
		})
	}
	update.IncludeSpreadsheetInResponse = true
	call := serv.api.Spreadsheets.BatchUpdate(group.SpreadsheetId, &update)
	resp, err := call.Context(ctx).Do()
	if err != nil {
		return err
	}
	_, err = serv.api.Spreadsheets.BatchUpdate(group.SpreadsheetId, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{DeleteSheet: &sheets.DeleteSheetRequest{SheetId: resp.UpdatedSpreadsheet.Sheets[0].Properties.SheetId}},
		},
	},
	).Context(ctx).Do()
	if err != nil {
		return err
	}
	return nil
}

func (serv *SheetsApiService) createLessonName(lesson persistance.Lesson) string {
	updateTitle := lesson.Subject + " " + serv.formatDateToEuropean(lesson.Date)
	if iis_api_entities.Subgroup(lesson.SubgroupNumber) != iis_api_entities.AllSubgroups {
		updateTitle += fmt.Sprintf(" (%s)", fmt.Sprint(int(lesson.SubgroupNumber)))
	}
	return updateTitle
}

func (serv *SheetsApiService) formatDateToEuropean(date time.Time) string {
	zeroPrependedDay := fmt.Sprint(date.Day()/10) + fmt.Sprint(date.Day()%10)
	zeroPrependedMonth := fmt.Sprint(int(date.Month())/10) + fmt.Sprint(int(date.Month())%10)
	return zeroPrependedDay + "." + zeroPrependedMonth + "." + fmt.Sprint(date.Year())
}

func parseLessonName(name string) (subject string, date time.Time, subgroup iis_api_entities.Subgroup) {
	data := strings.Split(name, " ")
	subgroup = iis_api_entities.AllSubgroups
	if len(data) != 2 {
		if len(data) == 3 {
			subgroupNum, err := strconv.Atoi(data[2][1:2])
			if err != nil {
				return "", time.Time{}, iis_api_entities.AllSubgroups
			}
			subgroup = iis_api_entities.Subgroup(subgroupNum)
		} else {
			return "", time.Time{}, iis_api_entities.AllSubgroups
		}
	}
	subject = data[0]
	datePoints := strings.Split(data[1], ".")
	if len(datePoints) != 3 {
		return "", time.Time{}, iis_api_entities.AllSubgroups
	}
	day, err := strconv.Atoi(datePoints[0])
	if err != nil {
		return "", time.Time{}, iis_api_entities.AllSubgroups
	}
	month, err := strconv.Atoi(datePoints[1])
	if err != nil {
		return "", time.Time{}, iis_api_entities.AllSubgroups
	}
	year, err := strconv.Atoi(datePoints[2])
	if err != nil {
		return "", time.Time{}, iis_api_entities.AllSubgroups
	}
	date = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return subject, date, subgroup
}

func (serv *SheetsApiService) ClearSpreadsheet(ctx context.Context, spreadsheetId string, before time.Time) error {
	getSpreadsheetRequest := sheets.SpreadsheetsGetCall{}
	spreadsheet, err := getSpreadsheetRequest.Do()
	if err != nil {
		return err
	}
	deleteSheetsRequest := sheets.BatchUpdateSpreadsheetRequest{}
	for _, sheet := range spreadsheet.Sheets {
		_, date, _ := parseLessonName(sheet.Properties.Title)
		if date.Before(before) {
			deleteSheetsRequest.Requests = append(deleteSheetsRequest.Requests, &sheets.Request{
				DeleteSheet: &sheets.DeleteSheetRequest{SheetId: sheet.Properties.SheetId},
			})
		}
	}
	call := serv.api.Spreadsheets.BatchUpdate(spreadsheetId, &deleteSheetsRequest)
	_, err = call.Context(ctx).Do()
	return err
}

func (serv *SheetsApiService) AddLabworkRequest(ctx context.Context, req *labworks.AppendedLabwork) error {
	group, err := serv.groupsRepo.GetByName(ctx, req.GroupName)
	if err != nil {
		return err
	}
	spreadsheetId := group.SpreadsheetId
	spreadsheet, err := serv.api.Spreadsheets.Get(spreadsheetId).Context(ctx).Do()
	if err != nil {
		return err
	}
	for _, sheet := range spreadsheet.Sheets {
		titleSubject, titleDate, subgroupNum := parseLessonName(sheet.Properties.Title)
		if titleSubject == req.DisciplineName && req.RequestedDate.Equal(titleDate) && req.SubgroupNumber == subgroupNum {
			if sheet.Tables == nil {
				err = serv.createTable(ctx, spreadsheetId, sheet)
				if err != nil {
					return err
				}
			}
			err = serv.appendToSheet(ctx, spreadsheetId, sheet, req)
			return err
		}
	}
	return errors.New("no such labwork found")
}

func (serv *SheetsApiService) createTable(ctx context.Context, spreadsheetId string, sheet *sheets.Sheet) error {
	requests := []*sheets.Request{}
	for _, bandedRange := range sheet.BandedRanges {
		requests = append(requests, &sheets.Request{DeleteBanding: &sheets.DeleteBandingRequest{BandedRangeId: bandedRange.BandedRangeId}})
	}
	requests = append(requests, []*sheets.Request{
		{
			AddTable: &sheets.AddTableRequest{
				Table: &sheets.Table{
					Name: "Очередь",
					Range: &sheets.GridRange{
						SheetId:          sheet.Properties.SheetId,
						StartColumnIndex: 0,
						EndColumnIndex:   3,
					},
					RowsProperties: &sheets.TableRowsProperties{
						FirstBandColorStyle:  sheet.Properties.TabColorStyle,
						FooterColorStyle:     sheet.Properties.TabColorStyle,
						HeaderColorStyle:     sheet.Properties.TabColorStyle,
						SecondBandColorStyle: sheet.Properties.TabColorStyle,
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
	}...)
	_, err := serv.api.Spreadsheets.BatchUpdate(spreadsheetId, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to add table for sheet: %w", err)
	}
	return nil
}

func (serv *SheetsApiService) appendToSheet(ctx context.Context, spreadsheetId string, sheet *sheets.Sheet, req *labworks.AppendedLabwork) error {
	tableSearchRange := fmt.Sprintf("'%s'!A1:B5", sheet.Properties.Title)
	_, err := serv.api.Spreadsheets.Values.Append(spreadsheetId, tableSearchRange, &sheets.ValueRange{
		Range:          tableSearchRange,
		MajorDimension: "ROWS",
		Values:         [][]any{{req.FullName, req.LabworkNumber, serv.formatDateTimeToEuropean(req.SentProofTime)}},
	}).ValueInputOption("RAW").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to append values to sheet: %w", err)
	}
	_, err = serv.api.Spreadsheets.BatchUpdate(spreadsheetId, &sheets.BatchUpdateSpreadsheetRequest{Requests: []*sheets.Request{
		{
			SortRange: &sheets.SortRangeRequest{
				Range: &sheets.GridRange{SheetId: sheet.Properties.SheetId, StartRowIndex: 2},
				SortSpecs: []*sheets.SortSpec{
					{
						DimensionIndex: 2,
						SortOrder:      "DESCENDING",
					},
				},
			},
		},
	}}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to sort sheet after appending to it: %w", err)
	}
	return err
}

func (serv *SheetsApiService) formatDateTimeToEuropean(dateTime time.Time) string {
	zeroPrependedHour := fmt.Sprint(dateTime.Hour()/10) + fmt.Sprint(dateTime.Hour()%10)
	zeroPrependedMinutes := fmt.Sprint(dateTime.Minute()/10) + fmt.Sprint(dateTime.Minute()%10)
	zeroPrependedSeconds := fmt.Sprint(dateTime.Second()/10) + fmt.Sprint(dateTime.Second()%10)

	date := fmt.Sprint(dateTime.Day()) + "/" + fmt.Sprint(int(dateTime.Month())) + "/" + fmt.Sprint(dateTime.Year())
	time := zeroPrependedHour + ":" + zeroPrependedMinutes + ":" + zeroPrependedSeconds
	return date + " " + time
}

func (serv *SheetsApiService) Add(ctx context.Context, lesson *persistance.Lesson) error {
	group, err := serv.groupsRepo.GetById(ctx, int(lesson.GroupId))
	if err != nil {
		return fmt.Errorf("failed to get group in sheets api during addition of custom labwork: %w", err)
	}

	sheet, err := serv.api.Spreadsheets.Get(group.SpreadsheetId).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to get spreadsheet by id during addition of custom labwork: %w", err)
	}
	sheetIndex := 0
	sheetTitle := serv.createLessonName(*lesson)
	for i, sheet := range sheet.Sheets {
		if _, date, _ := parseLessonName(sheet.Properties.Title); date.After(lesson.Date) {
			sheetIndex = i + 1
			break
		}
	}

	if sheetIndex == 0 {
		sheetIndex = len(sheet.Sheets)
	}

	for _, sheet := range sheet.Sheets {
		if sheetTitle == sheet.Properties.Title {
			return errSheetExists
		}
	}

	createdSheet, err := serv.api.Spreadsheets.BatchUpdate(group.SpreadsheetId, &sheets.BatchUpdateSpreadsheetRequest{
		IncludeSpreadsheetInResponse: true,
		Requests: []*sheets.Request{
			{
				AddSheet: &sheets.AddSheetRequest{
					Properties: &sheets.SheetProperties{
						Title: sheetTitle,
						Index: int64(sheetIndex),
					},
				},
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to create sheet while adding custom labwork: %w", err)
	}

	err = serv.createTable(ctx, createdSheet.SpreadsheetId, createdSheet.UpdatedSpreadsheet.Sheets[sheetIndex])
	return err
}
