package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
)

const LESSONS_REQUESTS_TABLE = "lessons_requests"

var _ interfaces.LessonsRequestsRepository = (*LessonsRequestsRepository)(nil)

type LessonsRequestsRepository struct {
	interfaces.LessonsRequestsRepository
	db *sql.DB
}

func NewLessonsRequestsRepository(db *sql.DB) *LessonsRequestsRepository {
	return &LessonsRequestsRepository{
		db: db,
	}
}

func (repo *LessonsRequestsRepository) Add(ctx context.Context, req *entities.LessonRequest) error {
	query := fmt.Sprintf("INSERT INTO %s (user_id, lesson_id) values ($1, $2)", LESSONS_REQUESTS_TABLE)
	_, err := repo.db.ExecContext(ctx, query, req.UserId, req.LessonId)
	return err
}

func (repo *LessonsRequestsRepository) GetByUserId(ctx context.Context, userId int64) (*entities.LessonRequest, error) {
	query := fmt.Sprintf("SELECT id, user_id, group_id FROM %s WHERE user_id=$1", LESSONS_REQUESTS_TABLE)
	row := repo.db.QueryRowContext(ctx, query, userId)
	if row.Err() != nil {
		return nil, row.Err()
	}
	req := &entities.LessonRequest{}
	err := row.Scan(&req.Id, &req.UserId, &req.LessonId)
	if err != nil {
		return nil, err
	}
	return req, nil
}
