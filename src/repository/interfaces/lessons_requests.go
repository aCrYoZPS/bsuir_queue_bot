package interfaces

import (
	"context"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
)

type LessonsRequestsRepository interface {
	Add(context.Context,*entities.LessonRequest) error
	GetByUserId(ctx context.Context, userId int64) (*entities.LessonRequest, error)
}
