package sqlite

import (
	"database/sql"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
)

type GroupsRepository struct {
	interfaces.GroupsRepository
	db *sql.DB
}

func NewGroupsRepository(db *sql.DB) *GroupsRepository {
	return &GroupsRepository{
		db: db,
	}
}
