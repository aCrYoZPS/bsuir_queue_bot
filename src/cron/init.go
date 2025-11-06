package cron

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

type TasksRepository interface {
	Add(ctx context.Context, task PersistedTask) error
	GetCompleted(ctx context.Context, after time.Time) ([]PersistedTask, error)
}

type PersistedTask struct {
	ExecutedAt time.Time
	Name       string
}
type TasksController struct {
	sheets         SheetsApi
	lessons        LessonRepo
	drive          DriveApi
	lessonsRequest LessonsRequestRepo
	users          UsersRepo
	bot            *tgutils.Bot
	jobs           []gocron.Job
	tasksRepo      TasksRepository
}

func NewTasksController(sheets SheetsApi, lessons LessonRepo, lessonsRequest LessonsRequestRepo, users UsersRepo, drive DriveApi, tasks TasksRepository, bot *tgutils.Bot) *TasksController {
	tasksController := &TasksController{
		sheets:         sheets,
		lessons:        lessons,
		lessonsRequest: lessonsRequest,
		users:          users,
		drive:          drive,
		bot:            bot,
		tasksRepo:      tasks,
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

	daily := gocron.CronJob("17 16 * * *", false)

	sheetsRefresh := NewReminderTask(controller.sheets, controller.lessons, controller.lessonsRequest, controller.users, controller.bot)
	sheetsRefreshJob, err := scheduler.NewJob(daily, gocron.NewTask(func() { sheetsRefresh.Run(ctx) }), gocron.WithName("sheets refresh"), gocron.WithContext(ctx),
		gocron.WithEventListeners(gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
			err = controller.tasksRepo.Add(ctx, PersistedTask{ExecutedAt: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 22, 0, 0, 0, time.Local), Name: jobName})
			if err != nil {
				slog.Error("failed to add task: %s to db, error: %v", jobName, err)
			}
		})))
	if err != nil {
		slog.Error(fmt.Errorf("failed to init sheets refresh cron: %w", err).Error())
	}
	controller.jobs = append(controller.jobs, sheetsRefreshJob)

	clearLessons := NewClearLessonsTask(controller.sheets, controller.lessons, controller.drive)
	clearLessonsJob, err := scheduler.NewJob(daily, gocron.NewTask(func() { clearLessons.Run(ctx) }), gocron.WithName("clear lessons"), gocron.WithContext(ctx),
		gocron.WithEventListeners(gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
			err = controller.tasksRepo.Add(ctx, PersistedTask{ExecutedAt: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 22, 0, 0, 0, time.Local), Name: jobName})
			if err != nil {
				slog.Error("failed to add task: %s to db, error: %v", jobName, err)
			}
		})))
	if err != nil {
		slog.Error(fmt.Errorf("failed to init sheets refresh cron: %w", err).Error())
	}
	controller.jobs = append(controller.jobs, clearLessonsJob)

	scheduler.Start()
	controller.TasksExec(ctx)
	<-ctx.Done()
	err = scheduler.Shutdown()
	if err != nil {
		slog.Error(fmt.Errorf("failed to shutdown cron scheduler: %w", err).Error())
	}
}

func (controller *TasksController) TasksExec(ctx context.Context) {
	tasks, err := controller.tasksRepo.GetCompleted(ctx, time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-1, 0, 0, 0, 0, time.Local))
	if err != nil {
		slog.Error(fmt.Sprintf("failed to get tasks in tasks exec: %v", err))
	}
	taskRunToday := false
	for _, task := range tasks {
		// I am sure there is better way to compare...
		if time.Date(task.ExecutedAt.Year(), task.ExecutedAt.Month(), task.ExecutedAt.Day(), 0, 0, 0, 0, time.Local).
			Sub(time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)) < time.Hour*24 {
			taskRunToday = true
		}
	}
	if !taskRunToday {
		for _, job := range controller.jobs {
			err := job.RunNow()
			if err != nil {
				slog.Error(fmt.Sprintf("failedd to run task: %s, error: %v", job.Name(), err))
			}
		}
	}
}
