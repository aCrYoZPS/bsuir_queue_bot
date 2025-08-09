package interfaces

import (
	"context"
)

type LessonsService interface {
	AddGroupLessons(ctx context.Context, groupName string) (url string, err error)
}