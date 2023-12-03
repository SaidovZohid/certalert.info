package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/SaidovZohid/certalert.info/pkg/logger"
	"github.com/SaidovZohid/certalert.info/storage/models"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type notificationRepo struct {
	db  *pgxpool.Pool
	log logger.Logger
}

func NewNotifications(db *pgxpool.Pool, log logger.Logger) models.NotificationStorageI {
	return &notificationRepo{
		db:  db,
		log: log,
	}
}

func (n *notificationRepo) CreateNotificationRow(ctx context.Context, userID int64) error {
	query := `
		INSERT INTO notifications (
			user_id
		) VALUES ($1)
	`
	if _, err := n.db.Exec(ctx, query, userID); err != nil {
		return err
	}

	return nil
}

func (n *notificationRepo) UpdateTheAlertIntegrations(ctx context.Context, userID int64, nameField string, value bool) error {
	query := fmt.Sprintf(`
		UPDATE notifications
		SET %v = $1 WHERE user_id = $2
	`, nameField)
	if _, err := n.db.Exec(ctx, query, value, userID); err != nil {
		return err
	}

	return nil
}

func (n *notificationRepo) UpdateTheLastAlertTime(ctx context.Context, userID int64) error {
	query := `
		UPDATE notifications
		SET last_alert_time = $1 WHERE user_id = $2
	`
	if _, err := n.db.Exec(ctx, query, time.Now(), userID); err != nil {
		return err
	}

	return nil
}

func (n *notificationRepo) GetNotificationRowByUserID(ctx context.Context, userID int64) (*models.Notification, error) {
	var notification models.Notification
	query := `SELECT 
		user_id,
		expiry_alerts,
		change_alerts,
		before,
		email_alert,
		telegram_alert,
		slack_alert,
		discord_alert,
		microsoft_team_alert,
		last_alert_time
	FROM notifications WHERE user_id=$1`
	err := n.db.QueryRow(ctx, query, userID).Scan(
		&notification.UserID,
		&notification.ExpiryAlerts,
		&notification.ChangeAlert,
		&notification.Before,
		&notification.EmailAlert,
		&notification.TelegramAlert,
		&notification.SlackAlert,
		&notification.DiscordAlert,
		&notification.MicrosoftTeamsAlert,
		&notification.LastAlertTime,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	return &notification, nil
}
