package storage

import (
	"github.com/SaidovZohid/certalert.info/pkg/logger"
	"github.com/SaidovZohid/certalert.info/storage/models"
	"github.com/SaidovZohid/certalert.info/storage/postgres"
	"github.com/jackc/pgx/v4/pgxpool"
)

type StorageI interface {
	User() models.UserStorageI
	Session() models.SessionStorageI
	Domain() models.DomainStorageI
	Integrations() models.IntegrationsStorageI
	Notifications() models.NotificationStorageI
}

type StoragePg struct {
	userRepo      models.UserStorageI
	sessionRepo   models.SessionStorageI
	domainRepo    models.DomainStorageI
	integrations  models.IntegrationsStorageI
	notifications models.NotificationStorageI
}

func NewStoragePg(db *pgxpool.Pool, log logger.Logger) StorageI {
	return &StoragePg{
		userRepo:      postgres.NewUser(db, log),
		sessionRepo:   postgres.NewSession(db, log),
		domainRepo:    postgres.NewDomain(db, log),
		integrations:  postgres.NewIntegrations(db, log),
		notifications: postgres.NewNotifications(db, log),
	}
}

func (s *StoragePg) User() models.UserStorageI {
	return s.userRepo
}

func (s *StoragePg) Session() models.SessionStorageI {
	return s.sessionRepo
}

func (s *StoragePg) Domain() models.DomainStorageI {
	return s.domainRepo
}

func (s *StoragePg) Integrations() models.IntegrationsStorageI {
	return s.integrations
}

func (s *StoragePg) Notifications() models.NotificationStorageI {
	return s.notifications
}
