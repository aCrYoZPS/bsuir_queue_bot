package sqlite

import (
	"database/sql"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
)

type LessonsRepository struct {
	interfaces.LessonsRepository
	db *sql.DB
}

func NewLessonsRepository(db *sql.DB) *LessonsRepository {
	return &LessonsRepository{
		db: db,
	}
}

