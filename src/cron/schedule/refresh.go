package schedule

import (
	"context"
	"fmt"
	iis_api_entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"log/slog"
)

type LessonsService interface {
	AddUnpresentedLessons(ctx context.Context, groupName string) error
}

type GroupsRepository interface {
	GetActive(ctx context.Context) ([]iis_api_entities.Group, error)
}

type RefreshScheduleTask struct {
	groups  GroupsRepository
	lessons LessonsService
}

func NewRefreshScheduleTask(groups GroupsRepository, lessons LessonsService) *RefreshScheduleTask {
	return &RefreshScheduleTask{groups: groups, lessons: lessons}
}

func (task *RefreshScheduleTask) Run(ctx context.Context) {
	slog.Info("Started schedule refresh cron")
	groups, err := task.groups.GetActive(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get groups in refresh schedule task: %s", err.Error()))
		return
	}
	for _, group := range groups {
		err := task.lessons.AddUnpresentedLessons(ctx, group.Name)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to add unpresented lessons during refresh schedule task for group %s: %s", group.Name, err.Error()))
		}
	}
}
