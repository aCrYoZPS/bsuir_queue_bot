package cron

import (
	"context"
	"fmt"
	"log/slog"
	"time"

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
	scheduler gocron.Scheduler
	jobs      []gocron.Job
	tasksRepo TasksRepository
}

func NewTasksController(tasks TasksRepository) *TasksController {
	scheduler, err := gocron.NewScheduler()
	gocron.WithLocation(time.Local)
	if err != nil {
		panic(fmt.Errorf("failed to init cron scheduler: %w", err).Error())
	}
	tasksController := &TasksController{
		scheduler: scheduler,
		tasksRepo: tasks,
	}
	return tasksController
}

func (controller *TasksController) AddTask(definition gocron.JobDefinition, task gocron.Task, options ...gocron.JobOption) {
	totalOptions := make([]gocron.JobOption, 0)
	totalOptions = append(totalOptions, gocron.WithEventListeners(
		gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			err := controller.tasksRepo.Add(ctx, PersistedTask{
				ExecutedAt: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().Hour(), 0, 0, 0, time.Local), Name: jobName})
			if err != nil {
				slog.Error("failed to add task: %s to db, error: %v", jobName, err)
			}
		}),
	))
	for _, option := range options {
		totalOptions = append(totalOptions, option)
	}
	job, err := controller.scheduler.NewJob(definition, task, totalOptions...)
	if err != nil {
		panic(err.Error())
	}
	controller.jobs = append(controller.jobs, job)
}

func (controller *TasksController) InitTasks(ctx context.Context) {
	gocron.WithContext(ctx)
	controller.scheduler.Start()
	controller.TasksExec(ctx)
	<-ctx.Done()
	err := controller.scheduler.Shutdown()
	if err != nil {
		slog.Error(fmt.Errorf("failed to shutdown cron scheduler: %w", err).Error())
	}
}

func (controller *TasksController) TasksExec(ctx context.Context) {
	tasks, err := controller.tasksRepo.GetCompleted(ctx,
		time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-1, 0, 0, 0, 0, time.Local))
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
				slog.Error(fmt.Sprintf("failed to run task: %s: %v", job.Name(), err))
			}
		}
	}
}
