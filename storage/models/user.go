package models

import (
	"context"
	"time"
)

type UserStorageI interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}

type User struct {
	ID        int64     `bson:"_id"`
	FirstName string    `bson:"first_name"`
	LastName  string    `bson:"last_name"`
	Email     string    `bson:"email"`
	Password  string    `bson:"password"`
	CreatedAt time.Time `bson:"created_at"`
}
