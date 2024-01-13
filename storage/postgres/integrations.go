package postgres

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/SaidovZohid/certalert.info/pkg/logger"
	"github.com/SaidovZohid/certalert.info/storage/models"
)

type integrationsRepo struct {
	db  *pgxpool.Pool
	log logger.Logger
}

func NewIntegrations(db *pgxpool.Pool, log logger.Logger) models.IntegrationsStorageI {
	return &integrationsRepo{
		db:  db,
		log: log,
	}
}

func (i *integrationsRepo) GetFromTelegramByTGID(ctx context.Context, TGUserID int64) (*models.TelegramUser, error) {
	var user models.TelegramUser

	query := `
		SELECT 
			id,
			user_id,
			chat_id,
			telegram_user_id,
			lang,
			step,
			created_at
		FROM telegram_users WHERE  telegram_user_id = $1 
	`

	err := i.db.QueryRow(
		ctx,
		query,
		TGUserID,
	).Scan(
		&user.ID,
		&user.UserID,
		&user.ChatID,
		&user.TelegramUserID,
		&user.Lang,
		&user.Step,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (i *integrationsRepo) GetFromTelegramByUserID(ctx context.Context, userID int64) (*models.TelegramUser, error) {
	var user models.TelegramUser

	query := `
		SELECT 
			id,
			user_id,
			chat_id,
			telegram_user_id,
			lang,
			step,
			created_at
		FROM telegram_users WHERE  user_id = $1 
	`

	err := i.db.QueryRow(
		ctx,
		query,
		userID,
	).Scan(
		&user.ID,
		&user.UserID,
		&user.ChatID,
		&user.TelegramUserID,
		&user.Lang,
		&user.Step,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (i *integrationsRepo) LinkTelegramAccountToWebsiteAccount(ctx context.Context, user *models.TelegramUser) error {
	query := `
		INSERT INTO telegram_users (
			user_id,
			chat_id,
			telegram_user_id,
			lang,
			step
		) VALUES ($1, $2, $3, $4, $5)
	`
	_, err := i.db.Exec(ctx, query, user.UserID, user.ChatID, user.TelegramUserID, user.Lang, user.Step)
	if err != nil {
		return err
	}

	return nil
}
