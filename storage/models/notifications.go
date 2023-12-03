package models

import (
	"context"
)

type NotificationStorageI interface {
	CreateNotificationRow(ctx context.Context, userID int64) error
	UpdateTheAlertIntegrations(ctx context.Context, userID int64, nameField string, value bool) error
	GetNotificationRowByUserID(ctx context.Context, userID int64) (*Notification, error)
}

type Notification struct {
	UserID              int64
	ExpiryAlerts        bool // default true in db
	ChangeAlert         bool // default true in db
	Before              int
	EmailAlert          bool // default true in db
	TelegramAlert       bool // false
	SlackAlert          bool // false
	DiscordAlert        bool // false
	MicrosoftTeamsAlert bool // false
}
