package postgres

import (
	"context"

	"github.com/SaidovZohid/certalert.info/pkg/logger"
	"github.com/SaidovZohid/certalert.info/pkg/ssl"
	"github.com/SaidovZohid/certalert.info/storage/models"
	"github.com/jackc/pgx/v4/pgxpool"
)

type domainRepo struct {
	db  *pgxpool.Pool
	log logger.Logger
}

func NewDomain(db *pgxpool.Pool, log logger.Logger) models.DomainStorageI {
	return &domainRepo{
		db:  db,
		log: log,
	}
}

func (d *domainRepo) CreateTrackingDomain(ctx context.Context, domainInfo *ssl.DomainTracking) (*ssl.DomainTracking, error) {
	query := `
		INSERT INTO tracking_domains (
			domain,
			user_id,
			remote_address,
			issuer,
			signature_algo,
			public_key_algo,
			encoded_pem,
			public_key,
			signature,
			dns_names,
			key_usage,
			ext_key_usages,
			expires,
			status,
			last_poll_at,
			latency,
			error, 
			issued
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18) RETURNING id
	`
	err := d.db.QueryRow(
		ctx,
		query,
		domainInfo.DomainName,
		domainInfo.UserID,
		domainInfo.RemoteAddr,
		domainInfo.Issuer,
		domainInfo.SignatureAlgo,
		domainInfo.PublicKeyAlgo,
		domainInfo.EncodedPEM,
		domainInfo.PublicKey,
		domainInfo.Signature,
		domainInfo.DNSNames,
		domainInfo.KeyUsage,
		domainInfo.ExtKeyUsages,
		domainInfo.Expires,
		domainInfo.Status,
		domainInfo.LastPollAt,
		domainInfo.Latency,
		domainInfo.Error,
		domainInfo.Issued,
	).Scan(
		&domainInfo.ID,
	)
	if err != nil {
		return nil, err
	}

	return domainInfo, nil
}

func (d *domainRepo) GetDomainWithUserIDAndDomainName(ctx context.Context, domain *ssl.DomainTracking) (*ssl.DomainTracking, error) {
	query := `
		SELECT 
			id,
			remote_address,
			issuer,
			signature_algo,
			public_key_algo,
			encoded_pem,
			public_key,
			signature,
			dns_names,
			key_usage,
			ext_key_usages,
			issued,
			expires,
			status,
			last_poll_at,
			latency,
			error
		FROM tracking_domains WHERE user_id=$1 AND domain=$2
	`
	err := d.db.QueryRow(ctx, query, domain.UserID, domain.DomainName).Scan(
		&domain.ID,
		&domain.RemoteAddr,
		&domain.Issuer,
		&domain.SignatureAlgo,
		&domain.PublicKeyAlgo,
		&domain.EncodedPEM,
		&domain.PublicKey,
		&domain.Signature,
		&domain.DNSNames,
		&domain.KeyUsage,
		&domain.ExtKeyUsages,
		&domain.Issued,
		&domain.Expires,
		&domain.Status,
		&domain.LastPollAt,
		&domain.Latency,
		&domain.Error,
	)
	if err != nil {
		return nil, err
	}

	return domain, nil
}

func (d *domainRepo) GetDomainsWithUserID(ctx context.Context, userId int64) ([]*ssl.DomainTracking, error) {
	query := `
		SELECT 
			id,
			domain,
			remote_address,
			issuer,
			signature_algo,
			public_key_algo,
			encoded_pem,
			public_key,
			signature,
			dns_names,
			key_usage,
			ext_key_usages,
			issued,
			expires,
			status,
			last_poll_at,
			latency,
			error
		FROM tracking_domains
		WHERE user_id = $1
	`
	res, err := d.db.Query(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	response := make([]*ssl.DomainTracking, 0)
	for res.Next() {
		var domainInfo ssl.DomainTracking
		err := res.Scan(
			&domainInfo.ID,
			&domainInfo.DomainName,
			&domainInfo.RemoteAddr,
			&domainInfo.Issuer,
			&domainInfo.SignatureAlgo,
			&domainInfo.PublicKeyAlgo,
			&domainInfo.EncodedPEM,
			&domainInfo.PublicKey,
			&domainInfo.Signature,
			&domainInfo.DNSNames,
			&domainInfo.KeyUsage,
			&domainInfo.ExtKeyUsages,
			&domainInfo.Issued,
			&domainInfo.Expires,
			&domainInfo.Status,
			&domainInfo.LastPollAt,
			&domainInfo.Latency,
			&domainInfo.Error,
		)
		if err != nil {
			d.log.Error(err)
			continue
		}
		response = append(response, &domainInfo)
	}

	return response, nil
}

func (d *domainRepo) DeleteTrackingDomains(ctx context.Context, userID int64, domainsId []string) error {
	query := "DELETE FROM tracking_domains WHERE user_id = $1 AND id = $2"

	for _, v := range domainsId {
		_, err := d.db.Exec(ctx, query, userID, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *domainRepo) DeleteTrackingDomain(ctx context.Context, userID int64, domainId int64) error {
	query := "DELETE FROM tracking_domains WHERE user_id = $1 AND id = $2"

	_, err := d.db.Exec(ctx, query, userID, domainId)
	if err != nil {
		return err
	}

	return nil
}

func (d *domainRepo) UpdateExistingDomainInfo(ctx context.Context, domainInfo *ssl.DomainTracking) error {
	query := `UPDATE tracking_domains SET 
		remote_address = $1,
		issuer = $2,
		signature_algo = $3,
		public_key_algo = $4,
		encoded_pem = $5,
		public_key = $6,
		signature = $7,
		dns_names = $8,
		key_usage = $9,
		ext_key_usages = $10,
		expires = $11,
		status = $12,
		last_poll_at = $13,
		latency = $14,
		error = $15,
		issued = $16
	WHERE user_id = $17 AND id = $18
	`
	_, err := d.db.Exec(ctx, query, domainInfo.RemoteAddr, domainInfo.Issuer, domainInfo.SignatureAlgo, domainInfo.PublicKeyAlgo, domainInfo.EncodedPEM, domainInfo.PublicKey, domainInfo.Signature, domainInfo.DNSNames, domainInfo.KeyUsage, domainInfo.ExtKeyUsages, domainInfo.Expires, domainInfo.Status, domainInfo.LastPollAt, domainInfo.Latency, domainInfo.Error, domainInfo.Issued, domainInfo.UserID, domainInfo.ID)
	if err != nil {
		return err
	}

	return nil
}

func (d *domainRepo) GetDomainWithUserIDAndDomainID(ctx context.Context, userID int64, domainID int64) (*ssl.DomainTracking, error) {
	var domain ssl.DomainTracking
	query := `
		SELECT 
			id,
			domain,
			remote_address,
			issuer,
			signature_algo,
			public_key_algo,
			encoded_pem,
			public_key,
			signature,
			dns_names,
			key_usage,
			ext_key_usages,
			issued,
			expires,
			status,
			last_poll_at,
			latency,
			error
		FROM tracking_domains WHERE user_id=$1 AND id=$2
	`
	err := d.db.QueryRow(ctx, query, userID, domainID).Scan(
		&domain.ID,
		&domain.DomainName,
		&domain.RemoteAddr,
		&domain.Issuer,
		&domain.SignatureAlgo,
		&domain.PublicKeyAlgo,
		&domain.EncodedPEM,
		&domain.PublicKey,
		&domain.Signature,
		&domain.DNSNames,
		&domain.KeyUsage,
		&domain.ExtKeyUsages,
		&domain.Issued,
		&domain.Expires,
		&domain.Status,
		&domain.LastPollAt,
		&domain.Latency,
		&domain.Error,
	)
	if err != nil {
		return nil, err
	}

	return &domain, nil
}
