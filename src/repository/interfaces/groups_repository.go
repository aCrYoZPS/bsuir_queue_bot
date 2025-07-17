package interfaces

import (
	entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
)

type GroupsRepository interface {
	GetById(id int) (*entities.Group, error)
	GetAll() ([]entities.Group, error)
	Add(group *entities.Group) error
	AddRange(groups []entities.Group) error
	Update(group *entities.Group) error
	Delete(id int) error
}
