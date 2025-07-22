package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
)

var _ interfaces.AdminRequestsRepository = (*AdminRequestsRepository)(nil)

const ADMIN_REQUESTS_TABLE = "admin_requests"

type AdminRequestsRepository struct {
	db *sql.DB
}

func NewAdminRequestsRepository(db *sql.DB) *AdminRequestsRepository {
	return &AdminRequestsRepository{db: db}
}

func (repo *AdminRequestsRepository) SaveRequest(req *interfaces.AdminRequest) error {
	query := fmt.Sprintf("INSERT INTO %s (uuid, msg_id, chat_id) VALUES ($1, $2,$3)", ADMIN_REQUESTS_TABLE)
	_, err := repo.db.Exec(query, req.UUID, req.MsgId, req.ChatId)
	return err
}

func (repo *AdminRequestsRepository) DeleteRequest(msgId int64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE msg_id=$1", ADMIN_REQUESTS_TABLE)
	_, err := repo.db.Exec(query, msgId)
	return err
}

func (repo *AdminRequestsRepository) GetByUUID(uuid string) ([]interfaces.AdminRequest, error) {
	query := fmt.Sprintf("SELECT msg_id, chat_id FROM %s WHERE uuid=$1", ADMIN_REQUESTS_TABLE)
	rows, err := repo.db.Query(query, uuid)
	if err != nil {
		return nil, err
	}
	requests := []interfaces.AdminRequest{}
	for rows.Next() {
		if rows.Err() != nil {
			return nil, err
		}
		request := interfaces.AdminRequest{}
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

func (repo *AdminRequestsRepository) GetByMsg(msgId, chatId int64) (*interfaces.AdminRequest, error) {
	query := fmt.Sprintf("SELECT uuid, chat_id FROM %s WHERE msg_id=$1 and chat_id=$2", ADMIN_REQUESTS_TABLE)
	row := repo.db.QueryRow(query, msgId, chatId)
	if row.Err() != nil {
		return nil, row.Err()
	}
	request := &interfaces.AdminRequest{}
	err := row.Scan(&request.UUID, &request.ChatId)
	if err != nil {
		return nil, err
	}
	return request, nil
}
