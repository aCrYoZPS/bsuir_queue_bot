package interfaces

import (
	"context"
	"time"

	entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistence"
)

type LessonsRepository interface {
	GetNext(ctx context.Context, subject string, groupId int64) ([]persistence.Lesson, error)
	GetAll(ctx context.Context, groupName string) ([]persistence.Lesson, error)
	AddRange(context.Context, []*entities.Lesson) error
	Add(context.Context, *persistence.Lesson) error
	DeleteLessons(context.Context, time.Time) error
	GetEndedLessons(context.Context, time.Time) ([]persistence.Lesson, error)
	GetLessonByRequest(ctx context.Context, requestId int64) (*persistence.Lesson, error)
	GetSubjects(ctx context.Context, groupId int64) ([]string, error)
}
