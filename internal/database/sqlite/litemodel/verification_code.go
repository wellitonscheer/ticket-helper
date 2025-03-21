package litemodel

import (
	"time"
)

type VerificationCode struct {
	Id         int
	Email      string
	Code       int
	Expires_at time.Time
}

func (v VerificationCode) IsValid() bool {
	return v.Expires_at.After(time.Now().UTC())
}
