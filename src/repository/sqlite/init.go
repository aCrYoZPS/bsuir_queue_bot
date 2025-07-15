package sqlite

import (
	"database/sql"
	"io"
	"os"
)


func DatabaseInit(db *sql.DB) error {
	queryFile, err := os.Open(os.Getenv("SQLITE_INIT_FILE"))
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
