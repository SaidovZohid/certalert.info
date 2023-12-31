package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/SaidovZohid/certalert.info/pkg/logger"
	"github.com/SaidovZohid/certalert.info/storage/models"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type sessionRepo struct {
	db  *pgxpool.Pool
	log logger.Logger
}

func NewSession(db *pgxpool.Pool, log logger.Logger) models.SessionStorageI {
	return &sessionRepo{
		db:  db,
		log: log,
	}
}

func (s *sessionRepo) Session(ctx context.Context, session *models.Session) (*models.Session, error) {
	ses, err := s.GetSession(context.Background(), session.UserId, session.IpAddress, session.UserAgent)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	// if ses is nil create new session
	if ses == nil {
		query := `
		INSERT INTO sessions (
			id,
			user_id,
			access_token,
			ip_address,
			user_agent,
			expires_at,
			city,
			region,
			country,
			timezone,
			last_login
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at
		`

		err = s.db.QueryRow(
			ctx,
			query,
			session.ID,
			session.UserId,
			session.AccessToken,
			session.IpAddress,
			session.UserAgent,
			session.ExpiresAt,
			session.City,
			session.Region,
			session.Country,
			session.Timezone,
			time.Now().In(time.FixedZone("GMT+5", 5*60*60)),
		).Scan(
			&session.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		return session, nil
	}
	// if not update existing session
	query := `
		UPDATE sessions SET
			id=$1,
			access_token = $2,
			expires_at = $3,
			last_login = timezone('Asia/Tashkent', CURRENT_TIMESTAMP)
		WHERE user_id = $4 AND ip_address = $5 AND user_agent=$6
	`
	_, err = s.db.Exec(
		ctx,
		query,
		session.ID,
		session.AccessToken,
		session.ExpiresAt,
		session.UserId,
		session.IpAddress,
		session.UserAgent,
	)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *sessionRepo) GetSession(ctx context.Context, userId int64, ipAddress, Device string) (*models.Session, error) {
	query := `
		SELECT
			id,
			user_id,
			access_token,
			ip_address,
			user_agent,
			is_blocked,
			expires_at,
			created_at,
			city,
			region,
			country,
			timezone
		FROM sessions 
		WHERE user_id = $1 AND ip_address = $2 AND user_agent = $3
	`

	var session models.Session
	err := s.db.QueryRow(
		ctx,
		query,
		userId,
		ipAddress,
		Device,
	).Scan(
		&session.ID,
		&session.UserId,
		&session.AccessToken,
		&session.IpAddress,
		&session.UserAgent,
		&session.IsBlocked,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.City,
		&session.Region,
		&session.Country,
		&session.Timezone,
	)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (s *sessionRepo) GetSessionInfoByID(ctx context.Context, id string) (*models.Session, error) {
	query := `
		SELECT
			id,
			user_id,
			access_token,
			ip_address,
			user_agent,
			is_blocked,
			expires_at,
			created_at,
			city,
			region,
			country,
			timezone
		FROM sessions 
		WHERE id=$1
	`

	var session models.Session
	err := s.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&session.ID,
		&session.UserId,
		&session.AccessToken,
		&session.IpAddress,
		&session.UserAgent,
		&session.IsBlocked,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.City,
		&session.Region,
		&session.Country,
		&session.Timezone,
	)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (s *sessionRepo) DeleteSessionByID(ctx context.Context, id string) error {
	query := `
		DELETE FROM sessions WHERE id = $1
	`
	_, err := s.db.Exec(ctx, query, id)
	return err
}

func (s *sessionRepo) DeleteSessionByUserID(ctx context.Context, id int64) error {
	query := `
		DELETE FROM sessions WHERE user_id = $1
	`
	_, err := s.db.Exec(ctx, query, id)
	return err
}
