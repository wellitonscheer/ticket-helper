package liteservi

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/database/sqlite/litemodel"
)

type VerificationCodeService struct {
	db                  *sql.DB
	appContext          context.AppContext
	verificCodeLifetime time.Duration
}

func NewVerificationCodeService(appContext context.AppContext) VerificationCodeService {
	return VerificationCodeService{
		db:                  appContext.Sqlite,
		appContext:          appContext,
		verificCodeLifetime: appContext.Config.Common.LoginCodeLifetimeSec,
	}
}

func (v VerificationCodeService) GetByEmailCode(email string, code int) (litemodel.VerificationCode, error) {
	var verification litemodel.VerificationCode

	findCodeStmt := "SELECT * FROM verification_code WHERE email = ? AND code = ?"
	err := v.db.QueryRow(findCodeStmt, email, code).Scan(&verification.Id, &verification.Email, &verification.Code, &verification.Expires_at)
	if err != nil {
		if err == sql.ErrNoRows {
			return verification, fmt.Errorf("verification code not found")
		}

		return verification, fmt.Errorf("failed to select verification code: %w: email: %s: code %d", err.Error(), findCodeStmt, email, code)
	}

	return verification, nil
}

func (v VerificationCodeService) Add(verification litemodel.VerificationCode) error {
	insertCodeStmt := "INSERT INTO verification_code (email, code, expires_at) VALUES (?, ?, ?)"

	_, err := v.db.Exec(insertCodeStmt, verification.Email, verification.Code, verification.Expires_at)
	if err != nil {
		return fmt.Errorf("failed to insert verification code: %s: email%s: %v", err.Error(), verification.Email, verification.Expires_at)
	}

	return nil
}

func (v VerificationCodeService) NewVerificationCode(email string, code int) error {
	expAt := time.Now().UTC().Add(v.verificCodeLifetime)

	verification := litemodel.VerificationCode{
		Email:      email,
		Code:       code,
		Expires_at: expAt,
	}
	if err := v.Add(verification); err != nil {
		return fmt.Errorf("failed create verification code: %v: %s: %s: %d", err.Error(), email, code)
	}

	return nil
}
