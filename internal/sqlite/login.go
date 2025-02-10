package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
)

type liteLogin struct {
	db *sql.DB
}

func NewSqliteLogin() (*liteLogin, error) {
	db, err := sql.Open("sqlite3", "./ticket-helper.db")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sqlLite db: %v", err.Error())
	}

	return &liteLogin{
		db: db,
	}, nil
}

func (l *liteLogin) InsertAuthorizedEmails() error {
	defer l.db.Close()

	rawData, err := os.ReadFile("./data_source/authorized_emails.json")
	if err != nil {
		return fmt.Errorf("failed to read from json file: %v", err.Error())
	}

	var authorizedEmails []string
	err = json.Unmarshal(rawData, &authorizedEmails)
	if err != nil {
		return fmt.Errorf("failed to unmarshal authorized emails content: %v", err.Error())
	}

	sqlStmt := `
		CREATE TABLE IF NOT EXISTS authorized_emails (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL
		);
	`
	_, err = l.db.Exec(sqlStmt)
	if err != nil {
		return fmt.Errorf("failed to create authorized emails table: %v: %s", err.Error(), sqlStmt)
	}

	for _, email := range authorizedEmails {
		_, err = l.db.Exec("INSERT OR IGNORE INTO authorized_emails (email) VALUES (?)", email)
		if err != nil {
			return fmt.Errorf("failed to insert email: %v: %s", err.Error(), email)
		}
	}

	return nil
}

func (l *liteLogin) IsAuthorizedEmail() (bool, error) {
	defer l.db.Close()

	sqlStmt := `
	create table foo (id integer not null primary key, name text);
	`
	_, err := l.db.Exec(sqlStmt)
	if err != nil {
		return false, fmt.Errorf("failed to verify if authorized: %v: %s", err.Error(), sqlStmt)
	}

	return true, nil
}
