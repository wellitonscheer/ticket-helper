package sqlite

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/wellitonscheer/ticket-helper/internal/context"
)

type SqliteMigrations struct {
	appContext context.AppContext
}

func NewSqliteMigrations(appContext context.AppContext) SqliteMigrations {
	return SqliteMigrations{
		appContext: appContext,
	}
}

func (s SqliteMigrations) RunMigrations() {
	s.RunSessionMigration()
	s.RunAuthorizedEmailsMigration()
	s.RunVerificationCodeMigration()
}

func (s SqliteMigrations) RunSessionMigration() {
	createTbStmt := `
		CREATE TABLE IF NOT EXISTS session (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL,
			token TEXT NOT NULL,
			expires_at DATETIME NOT NULL
		);
	`

	if _, err := s.appContext.Sqlite.Exec(createTbStmt); err != nil {
		panic(err)
	}
}

func (s SqliteMigrations) RunAuthorizedEmailsMigration() {
	createTbStmt := `
		CREATE TABLE IF NOT EXISTS authorized_emails (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL
		);
	`
	_, err := s.appContext.Sqlite.Exec(createTbStmt)
	if err != nil {
		panic(fmt.Sprintf("failed to create authorized emails table: %v: sql: %s", err.Error(), createTbStmt))
	}

	rawData, err := os.ReadFile(s.appContext.Config.Data.AuthEmailsPath)
	if err != nil {
		panic(fmt.Sprintf("failed to read from json file: %v", err.Error()))
	}

	var authorizedEmails []string
	err = json.Unmarshal(rawData, &authorizedEmails)
	if err != nil {
		panic(fmt.Sprintf("failed to unmarshal authorized emails content: %v", err.Error()))
	}

	for _, email := range authorizedEmails {
		_, err = s.appContext.Sqlite.Exec("INSERT OR IGNORE INTO authorized_emails (email) VALUES (?)", email)
		if err != nil {
			panic(fmt.Sprintf("failed to insert email: %v: email used %s", err.Error(), email))
		}
	}
}

func (s SqliteMigrations) RunVerificationCodeMigration() {
	createTbStmt := `
		CREATE TABLE IF NOT EXISTS verification_code (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL,
			code INTEGER NOT NULL,
			expires_at DATETIME NOT NULL
		);
	`
	_, err := s.appContext.Sqlite.Exec(createTbStmt)
	if err != nil {
		panic(fmt.Sprintf("failed to create verification code table: %v", err.Error()))
	}
}
