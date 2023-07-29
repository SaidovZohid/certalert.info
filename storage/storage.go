package storage

import (
	"github.com/SaidovZohid/certalert.info/storage/models"
	"github.com/SaidovZohid/certalert.info/storage/postgres"
	"github.com/jmoiron/sqlx"
)

type StorageI interface {
	User() models.UserStorageI
	Session() models.SessionStorageI
}

type StoragePg struct {
	userRepo    models.UserStorageI
	sessionRepo models.SessionStorageI
}

func NewStoragePg(db *sqlx.DB) StorageI {
	return &StoragePg{
		userRepo:    postgres.NewUser(db),
		sessionRepo: postgres.NewSession(db),
	}
}

func (s *StoragePg) User() models.UserStorageI {
	return s.userRepo
}

func (s *StoragePg) Session() models.SessionStorageI {
	return s.sessionRepo
}
