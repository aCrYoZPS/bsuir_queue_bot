package iis_api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	sheetsapi "github.com/aCrYoZPS/bsuir_queue_bot/src/google_docs/sheets_api"
	iis_api_entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
)

// That'll probably be turned into service class, which contains injected repository. Commented it out to compile project, for now

type LessonsService struct {
	sheetsApi sheetsapi.SheetsApi
	interfaces.LessonsRepository
}

func NewLessonsService(repos interfaces.LessonsRepository, sheetsApi sheetsapi.SheetsApi) *LessonsService {
	return &LessonsService{
		LessonsRepository: repos,
		sheetsApi:         sheetsApi,
	}
}

type schedulesResponse struct {
	Monday    []*iis_api_entities.Lesson `json:"Понедельник" `
	Tuesday   []*iis_api_entities.Lesson `json:"Вторник"`
	Wednesday []*iis_api_entities.Lesson `json:"Среда"`
	Thursday  []*iis_api_entities.Lesson `json:"Четверг"`
	Friday    []*iis_api_entities.Lesson `json:"Пятница"`
	Saturday  []*iis_api_entities.Lesson `json:"Суббота"`
}

func (serv *LessonsService) AddGroupLessons(groupName string) (url string, err error) {
	responseJson, err := serv.getSchedulesJson(groupName)
	if err != nil {
		return "", err
	}
	lessons, err := serv.GetAll(groupName)
	if err != nil {
		return "", err
	}
	if len(lessons) != 0 {
		return serv.CreateFilledSheet(groupName, lessons)
	}
	totalLessons := serv.getTotalLessons(responseJson)
	err = serv.AddRange(totalLessons)
	if err != nil {
		return "", err
	}
	lessons, err = serv.GetAll(groupName)
	if err != nil {
		return "", err
	}
	return serv.CreateFilledSheet(groupName, lessons)
}

func (serv *LessonsService) getTotalLessons(responseJson *schedulesResponse) []*iis_api_entities.Lesson {
	return slices.Concat(responseJson.Monday, responseJson.Tuesday, responseJson.Wednesday, responseJson.Thursday, responseJson.Friday, responseJson.Saturday)
}

func (serv *LessonsService) CreateFilledSheet(groupName string, lessons []persistance.Lesson) (url string, err error) {
	url, err = serv.sheetsApi.CreateSheet(groupName)
	if err != nil {
		return "", err
	}
	err = serv.sheetsApi.CreateLists(groupName, lessons)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (serv *LessonsService) getSchedulesJson(groupName string) (*schedulesResponse, error) {
	iisApiUrl := fmt.Sprintf("https://iis.bsuir.by/api/v1/schedule?studentGroup=%s", groupName)
	client := &http.Client{}
	req, err := http.NewRequest("GET", iisApiUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	responseJson := &schedulesResponse{}
	groupId := int64(0)
	err = json.NewDecoder(resp.Body).Decode(&struct {
		GroupInfo struct{ Id *int64 } `json:"studentGroupDto"`
		Resp      *schedulesResponse  `json:"schedules"`
	}{struct{ Id *int64 }{&groupId}, responseJson})
	if err != nil {
		return nil, err
	}
	serv.assignGroupId(groupId, responseJson)
	return responseJson, nil
}

func (serv *LessonsService) assignGroupId(groupId int64, resp *schedulesResponse) {
	allLessons := slices.Concat(resp.Monday, resp.Tuesday, resp.Wednesday, resp.Thursday, resp.Friday, resp.Saturday)
	for i := range allLessons {
		allLessons[i].GroupId = groupId
	}
}
