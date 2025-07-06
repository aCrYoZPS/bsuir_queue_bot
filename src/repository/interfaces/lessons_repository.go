package interfaces

import entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"

type LessonsRepository interface {
	Add(lesson *entities.Lesson) error
	AddRange(lessons []entities.Lesson) error
	GetById(lessonId int) (*entities.Lesson, error)
	GetAllByGroupId(groupId int) ([]entities.Lesson, error)
	Update(lesson *entities.Lesson) error
	Delete(lessonId int) error
}
