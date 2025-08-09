package interfaces

import (
	"context"
	"time"

	entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
)

type LessonsRepository interface {
	GetNext(ctx context.Context, subject string, groupId int64) ([]persistance.Lesson, error)
	GetAll(ctx context.Context, groupName string) ([]persistance.Lesson, error)
	AddRange(context.Context, []*entities.Lesson) error
	Add(context.Context, *persistance.Lesson) error
	DeleteLessons(context.Context, time.Time) error
	GetEndedLessons(context.Context, time.Time) ([]persistance.Lesson, error)
	GetLessonByRequest(ctx context.Context, requestId int64) (*persistance.Lesson, error)
	GetSubjects(ctx context.Context, groupId int64) ([]string, error)
}
