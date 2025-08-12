package iis_api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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

func (serv *GroupsService) InitAllGroups(ctx context.Context) error {
	url := "https://iis.bsuir.by/api/v1/student-groups"

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	var data []entities.Group

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return fmt.Errorf("failed to decode data from response body to group entity in groups service: %w", err)
	}

	return serv.AddNonPresented(ctx, data)
}
