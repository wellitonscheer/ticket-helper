package sqlite

import (
	"database/sql"
	"fmt"
)

func ConectToSqliteDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./ticket-helper.db")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sqlLite db: %v", err.Error())
	}

	return db, nil
}
