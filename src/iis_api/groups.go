package iis_api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"

	entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
)

type GroupsService struct {
	interfaces.GroupsRepository
}

func NewGroupsService(repo interfaces.GroupsRepository) *GroupsService {
	return &GroupsService{
		GroupsRepository: repo,
	}
}

func (serv *GroupsService) InitAllGroups() error {
	if _, err := os.Stat("groups.json"); errors.Is(err, os.ErrNotExist) {
		file, err := os.Create("groups.json")
		if err != nil {
			return err
		}

		defer file.Close()

		url := "https://iis.bsuir.by/api/v1/student-groups"

		resp, err := http.Get(url)
		if err != nil {
			return err
		}

		defer resp.Body.Close()
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return err
		}

		slog.Info("Finished request")
	}

	file, err := os.Open("groups.json")
	if err != nil {
		return err
	}

	defer file.Close()

	var data []entities.Group

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return err
	}

	return serv.AddNonPresented(data)
}