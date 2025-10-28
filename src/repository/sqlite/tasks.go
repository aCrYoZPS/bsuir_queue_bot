package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/cron"
)

var _ cron.TasksRepository = (*TasksRepository)(nil)

const TASKS_TABLE = "tasks"

type TasksRepository struct {
	db *sql.DB
}

func NewTasksRepository(db *sql.DB) *TasksRepository {
	return &TasksRepository{db: db}
}

func (repo *TasksRepository) Add(ctx context.Context, task cron.PersistedTask) error {
	query := fmt.Sprintf("INSERT INTO %s VALUES (task_timestamp, task_name) VALUES ($1, $2)", TASKS_TABLE)
	_, err := repo.db.ExecContext(ctx, query, task.ExecutedAt.Unix(), task.Name)
	return err
}

func (repo *TasksRepository) GetCompleted(ctx context.Context, after time.Time) ([]cron.PersistedTask, error) {
	query := fmt.Sprintf("SELECT task_timestamp, task_name FROM %s WHERE task_timestamp>$1", TASKS_TABLE)
	rows, err := repo.db.QueryContext(ctx, query, after.Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []cron.PersistedTask{}
	for rows.Next() {
		task := cron.PersistedTask{}
		ExecutedAt := int64(0)
		err = rows.Scan(&ExecutedAt, &task.Name)
		if err != nil {
			return nil, err
		}
		task.ExecutedAt = time.Unix(ExecutedAt, 0)
		result = append(result, task)
	}
	return result, nil
}
