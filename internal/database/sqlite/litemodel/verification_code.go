package litemodel

import "time"

type VerificationCode struct {
	Id         int
	Email      string
	Code       int
	Expires_at time.Time
}
