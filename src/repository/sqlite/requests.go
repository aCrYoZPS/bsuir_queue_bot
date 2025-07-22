package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
)

var _ interfaces.RequestsRepository = (*RequestsRepository)(nil)

const REQUESTS_TABLE = "requests"

type RequestsRepository struct {
	db *sql.DB
}

func NewRequestsRepository(db *sql.DB) *RequestsRepository {
	return &RequestsRepository{db: db}
}

func (repo *RequestsRepository) SaveRequest(req *interfaces.Request) error {
	query := fmt.Sprintf("INSERT INTO %s (msg_id, chat_id) VALUES ($1, $2)", GROUPS_TABLE)
	_, err := repo.db.Exec(query, req.MsgID(), req.ChatId())
	return err
}

func (repo *RequestsRepository) DeleteRequest(req *interfaces.Request) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE msg_id=$1 and chat_id=$2", GROUPS_TABLE)
	_, err := repo.db.Exec(query, req.MsgID(), req.ChatId())
	return err
}
