package sheetsapi

import (
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
	groups, err := serv.groupsRepo.GetAllGroups()
	if err != nil {
		return err
	}

	for _, group := range groups {
		newSheet := sheets.Spreadsheet{Properties: &sheets.SpreadsheetProperties{
			Title: group.Name,
		}}

		serv.api.Spreadsheets.Create(&newSheet)
	}
	return nil
}

func (serv *SheetsApiService) createLists(groups []iis_api_entities.Group) error {
	for _, group := range groups {
		labworks, err := serv.lessonsRepo.GetAllLabworks(&group)
		if err != nil {
			return err
		}
		for _, labwork := range labworks {
			//TODO: create sheets for spreads
			req := sheets.BatchUpdateSpreadsheetRequest{Requests: []*sheets.Request{
				&sheets.Request{
					AddSheet: &sheets.AddSheetRequest{Properties: &sheets.SheetProperties{Title: ""}},
				}}}
			serv.api.Spreadsheets.BatchUpdate(group.SpreadsheetId, &req)
		}
	}
	return nil
}

func (serv *SheetsApiService) ClearLists() error {
	return nil
}

func ClearSheet() error {
	return nil
}
