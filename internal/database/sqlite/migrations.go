package sqlite

import (
	"database/sql"
)

type SqliteMigrations struct {
	db *sql.DB
}

func NewSqliteMigrations(db *sql.DB) SqliteMigrations {
	return SqliteMigrations{
		db: db,
	}
}

func (s SqliteMigrations) RunMigrations() {
	s.RunSessionMigration()
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

	if _, err := s.db.Exec(createTbStmt); err != nil {
		panic(err)
	}
}
