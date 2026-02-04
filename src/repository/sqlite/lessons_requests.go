package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"time"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
)

const (
	LESSONS_REQUESTS_TABLE = "lessons_requests"
	QUEUE_TABLE            = "queue"
)

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
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start tx during adding labwork request: %w", err)
	}
	defer tx.Rollback()
	query := fmt.Sprintf("INSERT INTO %s (user_id, lesson_id, msg_id, chat_id, submit_time, subgroup_num, is_pending) values ($1, $2, $3, $4, $5, $6, $7)", LESSONS_REQUESTS_TABLE)
	_, err = tx.ExecContext(ctx, query, req.UserId, req.LessonId, req.MsgId, req.ChatId, req.SubmitTime.Format(savedFormat), req.LabworkNumber, true)
	if err != nil {
		return fmt.Errorf("failed to insert request into table: %w", err)
	}
	err = repo.reorderRequestsTx(ctx, tx, req.LessonId)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit tx during addition of labwork request: %w", err)
	}
	return nil
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
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx during lsson request delition: %w", err)
	}
	var lessonId int64
	query := fmt.Sprintf("DELETE FROM %s WHERE id=$1 RETURNING lesson_id", LESSONS_REQUESTS_TABLE)
	row := tx.QueryRowContext(ctx, query, requestId)
	if row.Err() != nil {
		return fmt.Errorf("failed to delete lesson request: %w", err)
	}
	err = row.Scan(&lessonId)

	err = repo.reorderRequestsTx(ctx, tx, lessonId)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit deletion of lesson request: %w", err)
	}
	return nil
}

func (repo *LessonsRequestsRepository) SetToNextLesson(ctx context.Context, requestId int64) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	var lessonId int64
	query := fmt.Sprintf("UPDATE %s AS lr SET lesson_id = (SELECT id FROM lessons WHERE id>lr.lesson_id AND subject=(SELECT subject FROM %s WHERE id=(SELECT lesson_id FROM %[1]s WHERE id=$1))), resubmissions_count=resubmissions_count+1 WHERE id=$1 RETURNING lesson_id", LESSONS_REQUESTS_TABLE, LESSONS_TABLE)
	row := tx.QueryRowContext(ctx, query, requestId)
	if row.Err() != nil {
		return fmt.Errorf("failed to set to next lesson: %w", err)
	}
	err = row.Scan(&lessonId)
	if err != nil {
		return fmt.Errorf("failed to scan lesson id: %w", err)
	}
	err = repo.reorderRequestsTx(ctx, tx, lessonId)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (repo *LessonsRequestsRepository) SetAccepted(ctx context.Context, requestId int64) error {
	query := fmt.Sprintf("SELECT q.lesson_id, q.order_type, q.ascending FROM %s AS q INNER JOIN %s AS r ON r.lesson_id=$1 WHERE q.lesson_id=r.lesson_id", QUEUE_TABLE, LESSONS_REQUESTS_TABLE)
	rows, err := repo.db.QueryContext(ctx, query, requestId)
	if err != nil {
		return fmt.Errorf("failed to read lesson requests order: %w", err)
	}
	defer rows.Close()

	orderTypes := []persistance.OrderType{}
	var lessonId int64
	for rows.Next() {
		var orderType persistance.OrderType
		err = rows.Scan(&lessonId, &orderType.Value, &orderType.Ascending)
		if err != nil {
			return fmt.Errorf("failed to scan order types: %w", err)
		}
		orderTypes = append(orderTypes, orderType)
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	query = fmt.Sprintf("UPDATE %s SET is_pending=0 WHERE id=$1", LESSONS_REQUESTS_TABLE)
	_, err = tx.ExecContext(ctx, query, requestId)
	if err != nil {
		return fmt.Errorf("failed to set request as not pending: %w", err)
	}
	return nil
}

func (repo *LessonsRequestsRepository) ChangeOrderation(ctx context.Context, orderTypes []entities.OrderType, lessonId int64) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}

	defer tx.Rollback()

	query := fmt.Sprintf("DELETE FROM %s WHERE lesson_id=$1", QUEUE_TABLE)
	_, err = tx.ExecContext(ctx, query, lessonId)
	if err != nil {
		return fmt.Errorf("failed to delete previous queue sorting: %w", err)
	}

	err = repo.reorderRequestsTx(ctx, tx, lessonId)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (repo *LessonsRequestsRepository) reorderRequestsTx(ctx context.Context, tx *sql.Tx, lessonId int64) error {
	query := fmt.Sprintf("SELECT id, user_id, lesson_id, msg_id, chat_id, subgroup_num, submit_time FROM %s WHERE id=$1 AND is_pending=false", LESSONS_REQUESTS_TABLE)
	rows, err := tx.QueryContext(ctx, query, lessonId)
	if err != nil {
		return fmt.Errorf("failed to query requests for lesson: %w", err)
	}
	defer rows.Close()
	requests := []entities.LessonRequest{}
	for rows.Next() {
		var cur entities.LessonRequest
		err = rows.Scan(&cur.Id, &cur.UserId, &cur.LessonId, &cur.MsgId, &cur.ChatId, &cur.LabworkNumber, &cur.SubmitTime)
		if err != nil {
			return fmt.Errorf("failed to scan lesson request: %w", err)
		}
		requests = append(requests, cur)
	}

	orderTypes := make([]persistance.OrderType, 0)
	query = fmt.Sprintf("SELECT q.order_type, q.ascending FROM %s as q WHERE q.lesson_id=$1", QUEUE_TABLE)
	rows, err = tx.QueryContext(ctx, query, lessonId)
	if err != nil {
		return fmt.Errorf("failed to read lesson requests order during reordering: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cur persistance.OrderType
		err := rows.Scan(&cur.Value, &cur.Ascending)
		if err != nil {
			return fmt.Errorf("faield to scan row into order type: %w", err)
		}
		orderTypes = append(orderTypes, cur)
	}

	prev := func(_, _ entities.LessonRequest) int { return 0 }
	for _, orderType := range orderTypes {
		cur := func(_, _ entities.LessonRequest) int { return 0 }
		switch orderType.Value {
		case persistance.ByLabworkNumber:
			cur = func(a, b entities.LessonRequest) int {
				if orderType.Ascending {
					return int((a.LabworkNumber - b.LabworkNumber))
				} else {
					return (int(b.LabworkNumber - a.LabworkNumber))
				}
			}
		case persistance.BySubmission:
			cur = func(a, b entities.LessonRequest) int {
				if orderType.Ascending {
					return int(a.SubmitTime.Unix() - b.SubmitTime.Unix())
				} else {
					return int(a.SubmitTime.Unix() - b.SubmitTime.Unix())
				}
			}
		}
		slices.SortFunc(requests, func(a, b entities.LessonRequest) int {
			if prev(a, b) == 0 {
				return cur(a, b)
			}
			return 0
		})
	}

	query = fmt.Sprintf("UPDATE %s SET order_position=$1 WHERE id=$2", REQUESTS_TABLE)
	for i, request := range requests {
		_, err = tx.ExecContext(ctx, query, i+1, request.Id)
		if err != nil {
			return fmt.Errorf("failed to update query order: %w", err)
		}
	}
	return nil
}

func (repo *LessonsRequestsRepository) GetLabworkQueue(ctx context.Context, labworkId int64) ([]entities.User, error) {
	query := fmt.Sprintf("SELECT u.id, u.full_name, u.tg_id, u.group_id FROM %s AS l INNER JOIN %s as u ON u.tg_id=l.user_id WHERE l.lesson_id=$1 AND is_pending=TRUE ORDER BY order_position", LESSONS_REQUESTS_TABLE, USERS_TABLE)
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
