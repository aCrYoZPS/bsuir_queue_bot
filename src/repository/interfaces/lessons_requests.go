package interfaces

import (
	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
)

type LessonsRequestsRepository interface {
	Add(*entities.LessonRequest) error
	GetByUserId(userId int64) (*entities.LessonRequest, error)
}
