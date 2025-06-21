package interfaces

import iis_api_entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"

type GroupsRepository interface {
	GetAllGroups() ([]iis_api_entities.Group, error)
	SaveAllGroups([]iis_api_entities.Group) error
}

