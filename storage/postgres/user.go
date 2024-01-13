package postgres

import (
	"context"

	"github.com/SaidovZohid/certalert.info/pkg/logger"
	"github.com/SaidovZohid/certalert.info/storage/models"
	"github.com/jackc/pgx/v4/pgxpool"
)

type userRepo struct {
	db  *pgxpool.Pool
	log logger.Logger
}

func NewUser(db *pgxpool.Pool, log logger.Logger) models.UserStorageI {
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
			password,
		    sign_up_method,
			user_accepted_terms
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	err := u.db.QueryRow(
		ctx,
		query,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password,
		user.SignUpMethod,
		user.UserAcceptedTerms,
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
			max_domains_tracking,
			created_at
		FROM users WHERE email = $1
	`
	err := u.db.QueryRow(
		ctx,
		query,
		email,
	).Scan(
		&result.ID,
		&result.FirstName,
		&result.LastName,
		&result.Email,
		&result.Password,
		&result.LastPollAt,
		&result.MaxDomainsTracking,
		&result.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (u *userRepo) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
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
			max_domains_tracking,
			created_at
		FROM users WHERE id = $1
	`
	err := u.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&result.ID,
		&result.FirstName,
		&result.LastName,
		&result.Email,
		&result.Password,
		&result.LastPollAt,
		&result.MaxDomainsTracking,
		&result.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (u *userRepo) UpdateUserLastPoll(ctx context.Context, userID int64) error {
	query := `UPDATE users SET domains_last_check = CURRENT_TIMESTAMP WHERE id = $1`

	_, err := u.db.Exec(ctx, query, userID)
	return err
}

func (u *userRepo) UpdateUserLastPollToNULL(ctx context.Context, userID int64) error {
	query := `UPDATE users SET domains_last_check = NULL WHERE id = $1`

	_, err := u.db.Exec(ctx, query, userID)
	return err
}

func (u *userRepo) UpdateUserPassword(ctx context.Context, userID int64, newPassword string) error {
	query := `UPDATE users SET password = $1 WHERE id = $2`

	_, err := u.db.Exec(ctx, query, newPassword, userID)
	return err
}

func (u *userRepo) UpdateUserEmail(ctx context.Context, userID int64, newEmailAddr string) error {
	query := `UPDATE users SET email = $1 WHERE id = $2`

	_, err := u.db.Exec(ctx, query, newEmailAddr, userID)
	return err
}

func (u *userRepo) DeleteUser(ctx context.Context, userID int64) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := u.db.Exec(ctx, query, userID)
	return err
}
