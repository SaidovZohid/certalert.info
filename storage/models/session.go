package models

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SessionStorageI interface {
	Session(ctx context.Context, s *Session) (*Session, error)
	GetSessionInfoByID(ctx context.Context, id string) (*Session, error)
	DeleteSessionByID(ctx context.Context, id string) error
}

type Session struct {
	ID          uuid.UUID
	UserId      int64
	AccessToken string
	IpAddress   string
	UserAgent   string
	City        string
	Region      string
	Country     string
	Timezone    string
	LastLogin   string
	IsBlocked   bool
	ExpiresAt   time.Time
	CreatedAt   time.Time
}
