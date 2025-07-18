package iis_api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"

	entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
)

type GroupsService struct {
}

func NewGroupsService() *GroupsService {
	return &GroupsService{}
}

func (*GroupsService) GetAllGroups() ([]entities.Group, error) {
	// Later on it will be transfered to the db, however now that will do
	if _, err := os.Stat("groups.json"); errors.Is(err, os.ErrNotExist) {
		file, err := os.Create("groups.json")
		if err != nil {
			return nil, err
		}

		defer file.Close()

		url := "https://iis.bsuir.by/api/v1/student-groups"

		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return nil, err
		}

		slog.Info("Finished request")
	}

	file, err := os.Open("groups.json")
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var data []entities.Group

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func (*GroupsService) DoesGroupExist(name string) (bool, error) {
	return true, nil
}
