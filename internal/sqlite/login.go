package sqlite

import (
	"database/sql"
	"fmt"
	"time"
)

type LiteLogin struct {
	db *sql.DB
}

func NewSqliteLogin() (*LiteLogin, error) {
	db, err := sql.Open("sqlite3", "./ticket-helper.db")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sqlLite db: %v", err.Error())
	}

	return &LiteLogin{
		db: db,
	}, nil
}

func (l *LiteLogin) IsAuthorizedEmail(email string) (bool, error) {
	var authorized bool

	sqlStmt := "SELECT EXISTS(SELECT 1 FROM authorized_emails WHERE email = ?)"

	err := l.db.QueryRow(sqlStmt, email).Scan(&authorized)
	if err != nil {
		return false, fmt.Errorf("failed to verify if authorized: %v: %s: %s", err.Error(), sqlStmt, email)
	}

	return authorized, nil
}

func (l *LiteLogin) InsertVerificationCode(email string, code int) error {
	createTbStmt := `
		CREATE TABLE IF NOT EXISTS verification_code (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL,
			code INTEGER NOT NULL,
			expires_at DATETIME NOT NULL
		);
	`
	_, err := l.db.Exec(createTbStmt)
	if err != nil {
		return fmt.Errorf("failed to create verification code table: %v: %s", err.Error(), createTbStmt)
	}

	insertCodeStmt := "INSERT INTO verification_code (email, code, expires_at) VALUES (?, ?, ?)"

	_, err = l.db.Exec(insertCodeStmt, email, code, time.Now().Add(time.Minute*15))
	if err != nil {
		return fmt.Errorf("failed to insert verification code: %v: %s: %s: %d", err.Error(), insertCodeStmt, email, code)
	}

	return nil
}

type VerificationCode struct {
	Id         int
	Email      string
	Code       int
	Expires_at time.Time
}

func (l *LiteLogin) IsValidVefificationCode(email string, code int) (bool, error) {
	var verificationCode VerificationCode

	findCodeStmt := "SELECT * FROM verification_code WHERE email = ? AND code = ? AND expires_at >= DATETIME('now', 'localtime');"
	err := l.db.QueryRow(findCodeStmt, email, code).Scan(&verificationCode.Id, &verificationCode.Email, &verificationCode.Code, &verificationCode.Expires_at)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("invalid verification code")
		}

		return false, fmt.Errorf("failed to select verification code: %v: %s: %s: %d", err.Error(), findCodeStmt, email, code)
	}

	deleteCodeStmt := "DELETE FROM verification_code WHERE id = ?"
	_, err = l.db.Exec(deleteCodeStmt, verificationCode.Id)
	if err != nil {
		return false, fmt.Errorf("failed to delete used code: %v: %s: %s: %d", err.Error(), deleteCodeStmt, email, code)
	}

	fmt.Printf("verification: %+v", verificationCode)

	return true, nil
}

func (l *LiteLogin) CreateUserSession(email, token string) error {
	createTbStmt := `
		CREATE TABLE IF NOT EXISTS session (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL,
			token TEXT NOT NULL,
			expires_at DATETIME NOT NULL
		);
	`
	_, err := l.db.Exec(createTbStmt)
	if err != nil {
		return fmt.Errorf("failed to create session table: %v: %s", err.Error(), createTbStmt)
	}

	insertSessionStmt := "INSERT INTO session (email, token, expires_at) VALUES (?, ?, ?)"

	_, err = l.db.Exec(insertSessionStmt, email, token, time.Now().Add(time.Hour*3))
	if err != nil {
		return fmt.Errorf("failed to create session: %v: %s: %s: %s", err.Error(), insertSessionStmt, email, token)
	}

	return nil
}

type Session struct {
	Id         int
	Email      string
	Token      string
	Expires_at time.Time
}

func (l *LiteLogin) IsValidSession(token string) (bool, error) {
	var session Session

	findSessionStmt := "SELECT * FROM session WHERE token = ? AND expires_at >= DATETIME('now', 'localtime');"
	err := l.db.QueryRow(findSessionStmt, token).Scan(&session.Id, &session.Email, &session.Token, &session.Expires_at)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("invalid session")
		}

		return false, fmt.Errorf("failed to select session: %v: %s: %s", err.Error(), findSessionStmt, token)
	}

	return true, nil
}
