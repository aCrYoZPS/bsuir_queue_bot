package cron

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type spreadsheetId = string
type DriveApiClear interface {
	GetSpreadsheets(ctx context.Context) ([]spreadsheetId, error)
}
type SheetsApiClear interface {
	ClearSpreadsheet(ctx context.Context, spreadsheetId string, before time.Time) error
}
type LessonsRepoClear interface {
	DeleteLessons(context.Context, time.Time) error
}
type ClearLessonsTask struct {
	sheets  SheetsApiClear
	lessons LessonsRepoClear
	drive   DriveApiClear
}

func NewClearLessonsTask(sheets SheetsApiClear, lessons LessonsRepoClear, drive DriveApiClear) *ClearLessonsTask {
	return &ClearLessonsTask{sheets: sheets, lessons: lessons, drive: drive}
}

func (task *ClearLessonsTask) Run(ctx context.Context) {
	done := make(chan struct{}, 1)
	slog.Info("Started clear lessons cron")
	go func(chan struct{}) {
		defer func() { done<-struct{}{} }()
		spreadsheetIds, err := task.drive.GetSpreadsheets(ctx)
		if err != nil {
			slog.Error(fmt.Errorf("failed to get spreadsheets in clear lessons task: %w", err).Error())
		}
		deletionTime := time.Now().AddDate(0, 0, -14)
		for _, id := range spreadsheetIds {
			err = task.sheets.ClearSpreadsheet(ctx, id, deletionTime)
			if err != nil {
				slog.Error(fmt.Errorf("failed to clear spreadsheet id (%s) in clear lessons task: %w", id, err).Error())
			}
		}
		err = task.lessons.DeleteLessons(ctx, deletionTime)
		if err != nil {
			slog.Error(fmt.Errorf("failed to delete lessons in clear lesson task: %w", err).Error())
		}
	}(done)
	select {
	case <-ctx.Done():
		slog.Error(fmt.Errorf("failed to finish clear lessons task due to deadline: %w", ctx.Err()).Error())
	case <-done:
	}
}
