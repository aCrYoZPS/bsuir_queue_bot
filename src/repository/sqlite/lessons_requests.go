package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
)

const LESSONS_REQUESTS_TABLE = "lessons_requests"

var _ interfaces.LessonsRequestsRepository = (*LessonsRequestsRepository)(nil)

type LessonsRequestsRepository struct {
	db *sql.DB
}

func NewLessonsRequestsRepository(db *sql.DB) *LessonsRequestsRepository {
	return &LessonsRequestsRepository{
		db: db,
	}
}

func (repo *LessonsRequestsRepository) Add(ctx context.Context, req *entities.LessonRequest) error {
	query := fmt.Sprintf("INSERT INTO %s (user_id, lesson_id, msg_id, chat_id, submit_time) values ($1, $2, $3, $4, $5)", LESSONS_REQUESTS_TABLE)
	_, err := repo.db.ExecContext(ctx, query, req.UserId, req.LessonId, req.MsgId, req.ChatId, req.SubmitTime.Format(savedFormat))
	return err
}

func (repo *LessonsRequestsRepository) Get(ctx context.Context, id int64) (*entities.LessonRequest, error) {
	query := fmt.Sprintf("SELECT id, user_id, lesson_id, msg_id, chat_id, submit_time FROM %s WHERE id=$1", LESSONS_REQUESTS_TABLE)
	row := repo.db.QueryRowContext(ctx, query, id)
	if row.Err() != nil {
		return nil, row.Err()
	}
	req := &entities.LessonRequest{}
	var storedTime = ""
	err := row.Scan(&req.Id, &req.UserId, &req.LessonId, &req.MsgId, &req.ChatId, &storedTime)
	if err != nil {
		return nil, err
	}
	req.SubmitTime, _ = time.Parse(savedFormat, storedTime)
	return req, nil
}

func (repo *LessonsRequestsRepository) GetByUserId(ctx context.Context, userId int64) (*entities.LessonRequest, error) {
	query := fmt.Sprintf("SELECT id, user_id, chat_id,lesson_id FROM %s WHERE user_id=$1", LESSONS_REQUESTS_TABLE)
	row := repo.db.QueryRowContext(ctx, query, userId)
	if row.Err() != nil {
		return nil, row.Err()
	}
	req := &entities.LessonRequest{}
	err := row.Scan(&req.Id, &req.UserId, &req.ChatId, &req.LessonId)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (repo *LessonsRequestsRepository) GetLessonRequests(ctx context.Context, lessonId int64) ([]entities.LessonRequest, error) {
	query := fmt.Sprintf("SELECT id, user_id, lesson_id, msg_id, chat_id FROM %s WHERE lesson_id=$1", LESSONS_REQUESTS_TABLE)
	rows, err := repo.db.QueryContext(ctx, query, lessonId)
	if err != nil {
		return nil, err
	}
	requests := []entities.LessonRequest{}
	var req = &entities.LessonRequest{}
	for rows.Next() {
		if rows.Err() != nil {
			return nil, rows.Err()
		}
		err := rows.Scan(&req.Id, &req.UserId, &req.LessonId, &req.MsgId, &req.ChatId)
		if err != nil {
			return nil, err
		}
		requests = append(requests, *req)
	}
	return requests, nil
}

func (repo *LessonsRequestsRepository) Delete(ctx context.Context, requestId int64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id=$1", LESSONS_REQUESTS_TABLE)
	_, err := repo.db.ExecContext(ctx, query, requestId)
	return err
}

func (repo *LessonsRequestsRepository) SetToNextLesson(ctx context.Context, requestId int64) error {
	query := fmt.Sprintf("UPDATE %s AS lr SET lesson_id = (SELECT id FROM lessons WHERE id>lr.lesson_id AND subject=(SELECT subject FROM %s WHERE id=(SELECT lesson_id FROM %[1]s WHERE id=$1))) WHERE id=$1", LESSONS_REQUESTS_TABLE, LESSONS_TABLE)
	_, err := repo.db.ExecContext(ctx, query, requestId)
	return err
}
