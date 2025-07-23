package iis_api

import (
	"encoding/json"
	_ "encoding/json"
	"fmt"
	"net/http"
	_ "os"
	_ "time"

	sheetsapi "github.com/aCrYoZPS/bsuir_queue_bot/src/google_docs/sheets_api"
	_ "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	iis_api_entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
	_ "github.com/aCrYoZPS/bsuir_queue_bot/src/utils"
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
	Json map[string]any `json:"schedules" binding:"required"`
}

func (serv *LessonsService) AddGroupLessons(groupName string) (url string, err error) {
	responseJson, err := serv.getSchedulesJson(groupName)
	if err != nil {
		return "", err
	}
	totalLessons, err := serv.getLessonEntities(responseJson)
	if err != nil {
		return "", err
	}
	err = serv.AddRange(totalLessons)
	if err != nil {
		return "", err
	}
	lessons, err := serv.GetAll(groupName)
	if err != nil {
		return "", err
	}
	url, err = serv.CreateFilledSheet(groupName, lessons)
	if err != nil {
		return "", err
	}
	return url, nil
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
	err = json.NewDecoder(resp.Body).Decode(responseJson)
	if err != nil {
		return nil, err
	}
	return responseJson, nil
}

func (serv *LessonsService) getLessonEntities(responseJson *schedulesResponse) ([]iis_api_entities.Lesson, error) {
	totalLessons := []iis_api_entities.Lesson{}
	for _, value := range responseJson.Json {
		schedule, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		lessons := []iis_api_entities.Lesson{}
		err = json.Unmarshal(schedule, &lessons)
		if err != nil {
			return nil, err
		}
		totalLessons = append(totalLessons, lessons...)
	}
	return totalLessons, nil
}
