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

func (l *liteLogin) IsAuthorizedEmail(email string) (bool, error) {
	var authorized bool

	sqlStmt := "SELECT EXISTS(SELECT 1 FROM authorized_emails WHERE email = ?)"

	err := l.db.QueryRow(sqlStmt, email).Scan(&authorized)
	if err != nil {
		return false, fmt.Errorf("failed to verify if authorized: %v: %s: %s", err.Error(), sqlStmt, email)
	}

	return authorized, nil
}

func (l *liteLogin) InsertVerificationCode(email string, code int) error {
	createTbStmt := `
		CREATE TABLE IF NOT EXISTS verification_code (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL,
			code INTEGER NOT NULL
		);
	`
	_, err := l.db.Exec(createTbStmt)
	if err != nil {
		return fmt.Errorf("failed to create verification code table: %v: %s", err.Error(), createTbStmt)
	}

	insertCodeStmt := "INSERT INTO verification_code (email, code) VALUES (?, ?)"

	_, err = l.db.Exec(insertCodeStmt, email, code)
	if err != nil {
		return fmt.Errorf("failed to insert verification code: %v: %s: %s: %d", err.Error(), insertCodeStmt, email, code)
	}

	return nil
}
