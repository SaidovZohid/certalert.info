package models

import (
	"context"
	"time"
)

type IntegrationsStorageI interface {
	GetFromTelegramByTGID(ctx context.Context, TGUserID int64) (*TelegramUser, error)
	LinkTelegramAccountToWebsiteAccount(ctx context.Context, user *TelegramUser) error
	GetFromTelegramByUserID(ctx context.Context, userID int64) (*TelegramUser, error)
}

type TelegramUser struct {
	ID             int64
	UserID         int64
	ChatID         int64
	TelegramUserID int64
	Lang           string
	Step           string
	CreatedAt      time.Time
}
