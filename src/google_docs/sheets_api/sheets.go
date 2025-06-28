package sheetsapi

import (
	"fmt"
	"slices"
	"sort"
	"time"

	iis_api_entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/utils"
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
	//I haven't figured out a way to batch these requests :(
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

// Labwork has date of start. Here we have struct that simply represents a specific lesson
type Lesson struct {
	date     time.Time
	time     time.Time
	subject  string
	subgroup iis_api_entities.Subgroup
}

func EntityToLesson(labwork iis_api_entities.Lesson, date time.Time) *Lesson {
	return &Lesson{
		date,
		time.Time(labwork.StartTime),
		labwork.Subject,
		labwork.SubgroupNumber,
	}
}

func (serv *SheetsApiService) CreateLists() error {
	groups, err := serv.groupsRepo.GetAll()
	if err != nil {
		return err
	}
	for _, group := range groups {
		var update = sheets.BatchUpdateSpreadsheetRequest{}
		labworks, err := serv.lessonsRepo.GetAllLabworks(&group)
		if err != nil {
			return err
		}

		sortedLessons := serv.getSortedLessons(labworks)
		for _, lesson := range sortedLessons {
			updateTitle := lesson.subject + " " + serv.formatDateToEuropean(lesson.date)
			if lesson.subgroup != iis_api_entities.AllSubgroups {
				updateTitle += fmt.Sprintf(" (%s)", fmt.Sprint(int(lesson.subgroup)))
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

func (serv *SheetsApiService) getSortedLessons(labworks []iis_api_entities.Lesson) []Lesson {
	if len(labworks) == 0 {
		return nil
	}
	var lessons = make([]Lesson, 0, len(labworks)*10)
	lessons = append(lessons, *EntityToLesson(labworks[0], time.Time(labworks[0].StartDate)))
	for _, labwork := range labworks {
		var startDate, endDate = time.Time(labwork.StartDate), time.Time(labwork.EndDate)
		currentDate := startDate
		for !currentDate.Equal(endDate) {
			currentDate = currentDate.Add(time.Hour * 24 * 7 * serv.calculateWeeksDistance(labwork.WeekNumber, utils.CalculateWeek(startDate)))
			lessons = append(lessons, *EntityToLesson(labwork, currentDate))
		}
	}

	sort.Slice(lessons, func(i, j int) bool {
		if !lessons[i].date.Equal(lessons[j].date) {
			return lessons[i].date.Before(lessons[j].date)
		}
		return lessons[i].time.Before(lessons[j].time)
	})

	return lessons
}

type Week = int8

// Returns value from 0 to 3, to measure distance in weeks between labworks.
// Doesn't handle cases, where week is unpresent in slice of weeks
func (serv *SheetsApiService) calculateWeeksDistance(weeks []Week, current Week) time.Duration {
	return time.Duration(weeks[(slices.Index(weeks, current)+1)%len(weeks)] - current)
}

// 24.04.2005 format of date
func (serv *SheetsApiService) formatDateToEuropean(date time.Time) string {
	return fmt.Sprint(date.Day()) + "." + fmt.Sprint(date.Month()) + "." + fmt.Sprint(date.Year())
}

func (serv *SheetsApiService) ClearLists() error {
	return nil
}
