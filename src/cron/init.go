package cron

import (
	"context"
	"fmt"
	"log/slog"

	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	"github.com/go-co-op/gocron/v2"
)

type TasksController struct {
	sheets         SheetsApi
	lessons        LessonsRepo
	lessonsRequest LessonsRequestsRepository
	users          UsersRepo
	bot            *tgutils.Bot
}

func NewTasksController(sheets SheetsApi, lessons LessonsRepo, lessonsRequest LessonsRequestsRepository, users UsersRepo, bot *tgutils.Bot) *TasksController {
	tasksController := &TasksController{
		sheets:         sheets,
		lessons:        lessons,
		lessonsRequest: lessonsRequest,
		users:          users,
		bot:            bot,
	}
	return tasksController
}

func (controller *TasksController) InitTasks(ctx context.Context) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		slog.Error(fmt.Errorf("failed to init cron scheduler: %w", err).Error())
	}
	sheetsRefresh := NewReminderTask(controller.sheets, controller.lessons, controller.lessonsRequest, controller.users, controller.bot)
	daily := gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(22, 0, 0)))
	_, err = scheduler.NewJob(daily, gocron.NewTask(func() { sheetsRefresh.Run() }))
	if err != nil {
		slog.Error(fmt.Errorf("failed to init sheets refresh cron: %w", err).Error())
	}

	scheduler.Start()
	<-ctx.Done()
	err = scheduler.Shutdown()
	if err != nil {
		slog.Error(fmt.Errorf("failed to shutdown cron scheduler: %w", err).Error())
	}
}
