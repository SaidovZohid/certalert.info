package models

import (
	"context"
	"time"
)

type UserStorageI interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUserLastPoll(ctx context.Context, userID int64) error
	UpdateUserLastPollToNULL(ctx context.Context, userID int64) error
	UpdateUserPassword(ctx context.Context, userID int64, newPassword string) error
}

type User struct {
	ID                 int64
	FirstName          string
	LastName           string
	Email              string
	Password           string
	LastPollAt         *time.Time
	MaxDomainsTracking *int
	UserAcceptedTerms  *bool
	CreatedAt          time.Time
}
