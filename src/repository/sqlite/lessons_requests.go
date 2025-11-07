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
	query := fmt.Sprintf("INSERT INTO %s (user_id, lesson_id, msg_id, chat_id, submit_time, subgroup_num, is_pending) values ($1, $2, $3, $4, $5, $6, $7)", LESSONS_REQUESTS_TABLE)
	_, err := repo.db.ExecContext(ctx, query, req.UserId, req.LessonId, req.MsgId, req.ChatId, req.SubmitTime.Format(savedFormat), req.LabworkNumber, true)
	return err
}

func (repo *LessonsRequestsRepository) Get(ctx context.Context, id int64) (*entities.LessonRequest, error) {
	query := fmt.Sprintf("SELECT id, user_id, lesson_id, msg_id, chat_id, subgroup_num, submit_time FROM %s WHERE id=$1", LESSONS_REQUESTS_TABLE)
	row := repo.db.QueryRowContext(ctx, query, id)
	if row.Err() != nil {
		return nil, row.Err()
	}
	req := &entities.LessonRequest{}
	var storedTime = ""
	err := row.Scan(&req.Id, &req.UserId, &req.LessonId, &req.MsgId, &req.ChatId, &req.LabworkNumber, &storedTime)
	if err != nil {
		return nil, err
	}
	req.SubmitTime, _ = time.Parse(savedFormat, storedTime)
	return req, nil
}

func (repo *LessonsRequestsRepository) GetByTgIds(ctx context.Context, msgId int64, chatId int64) (*entities.LessonRequest, error) {
	query := fmt.Sprintf("SELECT id, user_id, chat_id,lesson_id, msg_id, subgroup_num, submit_time FROM %s WHERE msg_id=$1 AND chat_id=$2", LESSONS_REQUESTS_TABLE)
	row := repo.db.QueryRowContext(ctx, query, msgId, chatId)
	if row.Err() != nil {
		return nil, row.Err()
	}
	req := &entities.LessonRequest{}
	var storedTime string
	err := row.Scan(&req.Id, &req.UserId, &req.ChatId, &req.LessonId, &req.MsgId, &req.LabworkNumber, &storedTime)
	req.SubmitTime, _ = time.Parse(time.RFC3339, storedTime)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (repo *LessonsRequestsRepository) GetLessonRequests(ctx context.Context, lessonId int64) ([]entities.LessonRequest, error) {
	query := fmt.Sprintf("SELECT id, user_id, lesson_id, msg_id, chat_id, subgroup_num FROM %s WHERE lesson_id=$1", LESSONS_REQUESTS_TABLE)
	rows, err := repo.db.QueryContext(ctx, query, lessonId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	requests := []entities.LessonRequest{}
	var req = &entities.LessonRequest{}
	for rows.Next() {
		if rows.Err() != nil {
			return nil, rows.Err()
		}
		err := rows.Scan(&req.Id, &req.UserId, &req.LessonId, &req.MsgId, &req.ChatId, &req.LabworkNumber)
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

func (repo *LessonsRequestsRepository) SetAccepted(ctx context.Context, requestId int64) error {
	query := fmt.Sprintf("UPDATE %s SET is_pending=FALSE WHERE id=$1", LESSONS_REQUESTS_TABLE)
	_, err := repo.db.ExecContext(ctx, query, requestId)
	return err
}

func (repo *LessonsRequestsRepository) GetLabworkQueue(ctx context.Context, labworkId int64) ([]entities.User, error) {
	query := fmt.Sprintf("SELECT u.id, u.full_name, u.tg_id, u.group_id FROM %s AS l INNER JOIN %s as u ON u.tg_id=l.user_id WHERE l.lesson_id=$1", LESSONS_REQUESTS_TABLE, USERS_TABLE)
	rows, err := repo.db.Query(query, labworkId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := make([]entities.User, 0, 10)
	var curUser entities.User
	for rows.Next() {
		err = rows.Scan(&curUser.Id, &curUser.FullName, &curUser.TgId, &curUser.GroupId)
		if err != nil {
			return nil, err
		}
		users = append(users, curUser)
	}
	return users, nil
}
