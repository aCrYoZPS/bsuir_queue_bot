package interfaces

import "github.com/aCrYoZPS/bsuir_queue_bot/src/entities"

type UsersRepository interface {
	GetById(id int64) (*entities.User, error)
	GetByTgId(tgId int64) (*entities.User, error)
	GetAll() ([]entities.User, error)
	Add(user *entities.User) error
	AddRange(users []entities.User) error
	Update(user *entities.User) error
	Delete(id int64) error
}
