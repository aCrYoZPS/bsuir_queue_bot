package interfaces

import (
	"context"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
)

type LessonsRequestsRepository interface {
	Add(context.Context, *entities.LessonRequest) error
	GetByTgIds(ctx context.Context, msgId int64, chatId int64) (*entities.LessonRequest, error)
	GetLessonRequests(ctx context.Context, lessonId int64) ([]entities.LessonRequest, error)
	SetAccepted(ctx context.Context, requestId int64) error
	Delete(ctx context.Context, requestId int64) error
}
