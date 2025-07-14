package litemodel

import "time"

type Session struct {
	Id         int
	Email      string
	Token      string
	Expires_at time.Time
}

func (s *Session) IsValid() bool {
	return s.Expires_at.After(time.Now().UTC())
}
