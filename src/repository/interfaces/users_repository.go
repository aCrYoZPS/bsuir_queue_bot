package interfaces

import (
	"context"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
)

type UsersRepository interface {
	GetById(ctx context.Context, id int64) (*entities.User, error)
	GetByTgId(ctx context.Context, tgId int64) (*entities.User, error)
	GetAll(ctx context.Context) ([]entities.User, error)
	Add(ctx context.Context, user *entities.User) error
	AddRange(ctx context.Context, users []entities.User) error
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id int64) error
}
