package liteservi

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/database/sqlite/litemodel"
)

type SessionService struct {
	db         *sql.DB
	appContext context.AppContext
}

func NewSessionService(appContext context.AppContext) SessionService {
	return SessionService{
		db:         appContext.Sqlite,
		appContext: appContext,
	}
}

func (s SessionService) GetByToken(token string) (litemodel.Session, error) {
	var session litemodel.Session

	sqlStmt := "SELECT * FROM session WHERE token = ?;"
	err := s.db.QueryRow(sqlStmt, token).Scan(&session)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return session, fmt.Errorf("no session found")
		}

		return session, fmt.Errorf("failed to select session: %v: %s: %s", err.Error(), sqlStmt, token)
	}

	return session, nil
}
