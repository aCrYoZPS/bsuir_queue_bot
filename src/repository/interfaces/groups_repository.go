package interfaces

import (
	"context"

	entities "github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	iisEntities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
)

type GroupsRepository interface {
	GetById(ctx context.Context, id int) (*iisEntities.Group, error)
	GetByName(ctx context.Context,name string) (*iisEntities.Group, error)
	GetAll(ctx context.Context) ([]iisEntities.Group, error)
	Add(ctx context.Context,group *iisEntities.Group) error
	AddRange(ctx context.Context,groups []iisEntities.Group) error
	AddNonPresented(ctx context.Context,groups []iisEntities.Group) error
	GetAdmins(ctx context.Context,groupName string) ([]entities.User, error)
	DoesGroupExist(ctx context.Context,groupName string) (bool, error)
	Update(ctx context.Context,group *iisEntities.Group) error
	Delete(ctx context.Context,id int) error
}
