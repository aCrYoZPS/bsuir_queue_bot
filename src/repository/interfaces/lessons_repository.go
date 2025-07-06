package interfaces

import (
	entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
)

type LessonsRepository interface {
	GetNextLabworks(subject string, groupId int64) ([]persistance.Lesson, error)
	GetAllLabworks(groupId int64) ([]persistance.Lesson, error)
	AddLabworks([]entities.Lesson) error
}
