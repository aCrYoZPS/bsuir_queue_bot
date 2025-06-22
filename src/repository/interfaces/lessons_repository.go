package interfaces

import entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"

type LessonsRepository interface {
	GetAllLabworks(*entities.Group) ([]entities.Lesson, error)
	GetNearestLabworks(group *entities.Group, subject string) ([]entities.Lesson, error)
}
