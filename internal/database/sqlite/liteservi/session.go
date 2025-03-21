package liteservi

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/database/sqlite/litemodel"
)

type SessionService struct {
	db              *sql.DB
	appContext      context.AppContext
	sessionLifetime time.Duration
}

func NewSessionService(appContext context.AppContext) SessionService {
	return SessionService{
		db:              appContext.Sqlite,
		appContext:      appContext,
		sessionLifetime: appContext.Config.Common.SessionLifetimeSec,
	}
}

func (s SessionService) GetByToken(token string) (litemodel.Session, error) {
	var session litemodel.Session

	sqlStmt := "SELECT * FROM session WHERE token = ?;"
	err := s.db.QueryRow(sqlStmt, token).Scan(&session.Id, &session.Email, &session.Token, &session.Expires_at)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return session, fmt.Errorf("no session found")
		}

		return session, fmt.Errorf("failed to select session: %w: %s: %s", err, sqlStmt, token)
	}

	return session, nil
}

func (s SessionService) Add(session litemodel.Session) error {
	insertSessionStmt := "INSERT INTO session (email, token, expires_at) VALUES (?, ?, ?)"

	_, err := s.db.Exec(insertSessionStmt, session.Email, session.Token, session.Expires_at)
	if err != nil {
		return fmt.Errorf("failed to insert session: %w: %+v", err, session)
	}

	return nil
}

func (s SessionService) NewSessionByEmail(email string) (string, error) {
	expAt := time.Now().UTC().Add(s.sessionLifetime)

	tokenUuid, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate uuid: %v", err)
	}
	tokenString := tokenUuid.String()

	session := litemodel.Session{
		Email:      email,
		Token:      tokenString,
		Expires_at: expAt,
	}

	if err = s.Add(session); err != nil {
		return "", fmt.Errorf("failed to insert new session: %v", err)
	}

	return tokenString, nil
}
