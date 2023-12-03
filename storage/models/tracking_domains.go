package models

import (
	"context"

	"github.com/SaidovZohid/certalert.info/pkg/ssl"
)

type DomainStorageI interface {
	CreateTrackingDomain(ctx context.Context, domainInfo *ssl.DomainTracking) (*ssl.DomainTracking, error)
	GetDomainsWithUserID(ctx context.Context, userId int64) ([]*ssl.DomainTracking, error)
	GetDomainWithUserIDAndDomainName(ctx context.Context, domain *ssl.DomainTracking) (*ssl.DomainTracking, error)
	GetDomainWithUserIDAndDomainID(ctx context.Context, userID int64, domainID int64) (*ssl.DomainTracking, error)
	DeleteTrackingDomains(ctx context.Context, userID int64, domainsID []string) error
	UpdateExistingDomainInfo(ctx context.Context, domainInfo *ssl.DomainTracking) error
	DeleteTrackingDomain(ctx context.Context, userID int64, domainId int64) error
	GetListofDomainsThatExists(ctx context.Context) ([]*ssl.DomainTracking, error)
	UpdateAllTheSameDomainsInfo(ctx context.Context, domainInfo *ssl.DomainTracking) error
	GetListofUsersThatDomainExists(ctx context.Context, domain string) ([]int64, error)
	UpdateTheLastAlertTime(ctx context.Context, userID int64, domain string) error
}
