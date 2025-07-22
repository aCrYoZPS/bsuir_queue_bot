package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
)

var _ interfaces.RequestsRepository = (*RequestsRepository)(nil)

const REQUESTS_TABLE = "group_requests"

type RequestsRepository struct {
	db *sql.DB
}

func NewRequestsRepository(db *sql.DB) *RequestsRepository {
	return &RequestsRepository{db: db}
}

func (repo *RequestsRepository) SaveRequest(req *interfaces.GroupRequest) error {
	query := fmt.Sprintf("INSERT INTO %s (uuid, msg_id, chat_id) VALUES ($1, $2,$3)", REQUESTS_TABLE)
	_, err := repo.db.Exec(query, req.UUID, req.MsgId, req.ChatId)
	return err
}

func (repo *RequestsRepository) DeleteRequest(msgId int64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE msg_id=$1", REQUESTS_TABLE)
	_, err := repo.db.Exec(query, msgId)
	return err
}

func (repo *RequestsRepository) GetByUUID(uuid string) ([]interfaces.GroupRequest, error) {
	query := fmt.Sprintf("SELECT msg_id, chat_id FROM %s WHERE uuid=$1", REQUESTS_TABLE)
	rows, err := repo.db.Query(query, uuid)
	if err != nil {
		return nil, err
	}
	requests := []interfaces.GroupRequest{}
	for rows.Next() {
		if rows.Err() != nil {
			return nil, err
		}
		request := interfaces.GroupRequest{}
		err := rows.Scan(&request.MsgId, &request.ChatId)
		if err != nil {
			return nil, err
		}
		request.UUID = uuid
		requests = append(requests, request)
	}
	if len(requests) == 0 {
		return nil, nil
	}
	return requests, nil
}

func (repo *RequestsRepository) GetByMsg(msgId, chatId int64) (*interfaces.GroupRequest, error) {
	query := fmt.Sprintf("SELECT uuid, chat_id FROM %s WHERE msg_id=$1 and chat_id=$2", REQUESTS_TABLE)
	row := repo.db.QueryRow(query, msgId, chatId)
	if row.Err() != nil {
		return nil, row.Err()
	}
	request := &interfaces.GroupRequest{}
	err := row.Scan(&request.UUID, &request.ChatId)
	if err != nil {
		return nil, err
	}
	return request, nil
}
