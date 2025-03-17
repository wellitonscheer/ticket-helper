package sqlite

import (
	"database/sql"
	"fmt"
)

func NewSqliteConnection() *sql.DB {
	fmt.Println("Connecting to sqlite now.")

	db, err := sql.Open("sqlite3", "./ticket-helper.db")
	if err != nil {
		panic(fmt.Errorf("failed to connect to sqlLite db: %v", err.Error()))
	}

	fmt.Println("Sqlite connected.")

	return db
}
