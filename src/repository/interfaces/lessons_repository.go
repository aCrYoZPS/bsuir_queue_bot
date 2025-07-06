package interfaces

import (
	entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
)

type LessonsRepository interface {
	GetNext(subject string, groupId int64) ([]persistance.Lesson, error)
	GetAll(groupId int64) ([]persistance.Lesson, error)
	AddRange([]entities.Lesson) error
}
