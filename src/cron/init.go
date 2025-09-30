package cron

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	"github.com/go-co-op/gocron/v2"
)

type TasksController struct {
	sheets         SheetsApi
	lessons        LessonRepo
	drive          DriveApi
	lessonsRequest LessonsRequestRepo
	users          UsersRepo
	bot            *tgutils.Bot
}

func NewTasksController(sheets SheetsApi, lessons LessonRepo, lessonsRequest LessonsRequestRepo, users UsersRepo, drive DriveApi, bot *tgutils.Bot) *TasksController {
	tasksController := &TasksController{
		sheets:         sheets,
		lessons:        lessons,
		lessonsRequest: lessonsRequest,
		users:          users,
		drive:          drive,
		bot:            bot,
	}
	return tasksController
}

type SheetsApi interface {
	SheetsApiClear
	SheetsApiReminder
}

type LessonRepo interface {
	LessonsRepoClear
	LessonsRepoReminder
}

type UsersRepo interface {
	UsersRepoReminder
}

type DriveApi interface {
	DriveApiClear
}
type LessonsRequestRepo interface {
	LessonsRequestsRepositoryReminder
}

func (controller *TasksController) InitTasks(ctx context.Context) {
	gocron.WithLocation(time.Local)
	gocron.WithContext(ctx)
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		slog.Error(fmt.Errorf("failed to init cron scheduler: %w", err).Error())
	}

	sheetsRefresh := NewReminderTask(controller.sheets, controller.lessons, controller.lessonsRequest, controller.users, controller.bot)
	daily := gocron.CronJob("55 23 * * *", false)

	_, err = scheduler.NewJob(daily, gocron.NewTask(func() { sheetsRefresh.Run(ctx) }), gocron.WithContext(ctx))
	if err != nil {
		slog.Error(fmt.Errorf("failed to init sheets refresh cron: %w", err).Error())
	}

	clearLessons := NewClearLessonsTask(controller.sheets, controller.lessons, controller.drive)
	_, err = scheduler.NewJob(daily, gocron.NewTask(func() { clearLessons.Run(ctx) }))
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
