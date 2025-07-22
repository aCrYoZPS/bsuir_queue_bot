package interfaces

import (
	entities "github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	iisEntities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
)

type GroupsRepository interface {
	GetById(id int) (*iisEntities.Group, error)
	GetByName(name string) (*iisEntities.Group, error)
	GetAll() ([]iisEntities.Group, error)
	Add(group *iisEntities.Group) error
	AddRange(groups []iisEntities.Group) error
	AddNonPresented(groups []iisEntities.Group) error
	GetAdmins(groupName string) ([]entities.User, error)
	DoesGroupExist(groupName string) (bool, error)
	Update(group *iisEntities.Group) error
	Delete(id int) error
}
