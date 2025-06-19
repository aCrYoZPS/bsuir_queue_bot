package iis_api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
)

func GetAllGroups() ([]entities.Group, error) {
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

	i := 0
	for _, group := range data {
		slog.Info(fmt.Sprintf("%+v", group))
		if i > 5 {
			break
		}
		i += 1
	}

	return data, nil
}
