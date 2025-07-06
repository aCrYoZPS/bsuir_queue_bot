package sqlite

import (
	"database/sql"
	"io"
	"os"
)

func DatabaseInit(db *sql.DB) error {
	queryFile, err := os.Open("queries/db_setup.sql")
	if err != nil {
		return err
	}
	query, err := io.ReadAll(queryFile)
	if err != nil {
		return err
	}
	_, err = db.Exec(string(query))
	return err
}
