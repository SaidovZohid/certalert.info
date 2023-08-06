package postgres

import (
	"context"

	"github.com/SaidovZohid/certalert.info/pkg/logger"
	"github.com/SaidovZohid/certalert.info/storage/models"
	"github.com/jmoiron/sqlx"
)

type userRepo struct {
	db  *sqlx.DB
	log logger.Logger
}

func NewUser(db *sqlx.DB, log logger.Logger) models.UserStorageI {
	return &userRepo{
		db:  db,
		log: log,
	}
}

func (u *userRepo) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
		INSERT INTO users (
			first_name,
			last_name,
			email,
			password
		) VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	err := u.db.QueryRow(
		query,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var (
		result models.User
	)

	query := `
		SELECT 
			id,
			first_name,
			last_name,
			email,
			password,
			domains_last_check,
			created_at
		FROM users WHERE email = $1
	`
	err := u.db.QueryRow(
		query,
		email,
	).Scan(
		&result.ID,
		&result.FirstName,
		&result.LastName,
		&result.Email,
		&result.Password,
		&result.LastPollAt,
		&result.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (d *userRepo) UpdateUserLastPoll(ctx context.Context, userID int64) error {
	query := `UPDATE users SET domains_last_check = CURRENT_TIMESTAMP WHERE id = $1`

	_, err := d.db.Exec(query, userID)
	return err
}
