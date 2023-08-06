package models

import (
	"context"

	"github.com/SaidovZohid/certalert.info/pkg/ssl"
)

type DomainStorageI interface {
	CreateTrackingDomain(ctx context.Context, domainInfo *ssl.DomainTracking) (*ssl.DomainTracking, error)
	GetDomainsWithUserID(ctx context.Context, userId int64) ([]*ssl.DomainTracking, error)
	GetDomainWithUserIDAndDomainName(ctx context.Context, domain *ssl.DomainTracking) (*ssl.DomainTracking, error)
	DeleteTrackingDomains(ctx context.Context, userID int64, domains []string) error
	UpdateExistingDomainInfo(ctx context.Context, domainInfo *ssl.DomainTracking) error
}
