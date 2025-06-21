package interfaces

import entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"

type LessonsRepository interface {
	GetAllLessons(*entities.Group) ([]entities.Lesson, error)
	GetNearestLessons(subject string) ([]entities.Lesson, error)
}
