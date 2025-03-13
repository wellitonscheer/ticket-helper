package liteservi

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/database/sqlite/litemodel"
)

type AuthorizedEmailsService struct {
	db         *sql.DB
	appContext context.AppContext
}

func NewAuthorizedEmailsService(appContext context.AppContext) AuthorizedEmailsService {
	return AuthorizedEmailsService{
		db:         appContext.Sqlite,
		appContext: appContext,
	}
}

func (a AuthorizedEmailsService) GetByEmail(email string) (litemodel.AuthorizedEmails, error) {
	var authEmail litemodel.AuthorizedEmails

	sqlStmt := "SELECT * FROM authorized_emails WHERE email = ?"

	err := a.db.QueryRow(sqlStmt, email).Scan(&authEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return authEmail, errors.New("no authorized email found")
		} else {
			return authEmail, fmt.Errorf("failed to get by email: %s", err.Error())
		}
	}

	return authEmail, nil
}

func (a AuthorizedEmailsService) IsAuthorizedEmail(email string) bool {
	if _, err := a.GetByEmail(email); err != nil {
		fmt.Printf("failed to verify if authorized: %w", err)
		return false
	}

	return true
}
