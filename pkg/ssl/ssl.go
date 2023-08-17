package ssl

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"
)

const (
	StatusInvalid      = "invalid"
	StatusOffline      = "offline"
	StatusHealthy      = "healthy"
	StatusExpires      = "expires"
	StatusExpired      = "expired"
	StatusUnResponsive = "unresponsive"
)

type TrackingDomainInfo struct {
	RemoteAddr    *string
	Issuer        *string
	SignatureAlgo *string
	PublicKeyAlgo *string
	EncodedPEM    *string
	PublicKey     *string
	Signature     *string
	DNSNames      *string
	KeyUsage      *string
	ExtKeyUsages  *string
	Issued        *time.Time
	Expires       *time.Time
	Status        *string
	LastPollAt    time.Time
	Latency       *int
	Error         *string
}

type DomainTracking struct {
	ID         int64
	UserID     int64
	DomainName string
	TrackingDomainInfo
}

func PollDomain(ctx context.Context, domain string) (*TrackingDomainInfo, error) {
	var (
		start    = time.Now()
		resultch = make(chan TrackingDomainInfo)
		config   = &tls.Config{}
		stv      string
		lt       int
	)

	go func() {
		conn, err := tls.Dial("tcp", fmt.Sprintf("%s:443", domain), config)
		if err != nil {
			stv = err.Error()
			lt = int(time.Since(start).Milliseconds())
			info := TrackingDomainInfo{
				LastPollAt: time.Now(),
				Error:      &stv,
				Latency:    &lt,
			}
			if isVerificationError(err) {
				stv = StatusInvalid
				info.Status = &stv
				resultch <- info
			}
			if isConnectionError(err) {
				stv = StatusOffline
				info.Status = &stv
				resultch <- info
			}

			return
		}
		defer conn.Close()
		var (
			state     = conn.ConnectionState()
			cert      = state.PeerCertificates[0]
			keyUsages = ""
		)

		for _, usage := range cert.ExtKeyUsage {
			keyUsages += " " + extKeyUsageToString(usage)
		}

		dnsNames := strings.Join(cert.DNSNames, ", ")
		org := cert.Issuer.Organization[0]
		lt = int(time.Since(start).Milliseconds())
		pubAlgo := cert.PublicKeyAlgorithm.String()
		sigAlgo := cert.SignatureAlgorithm.String()
		rmtAddr := conn.RemoteAddr().String()
		resultch <- TrackingDomainInfo{
			RemoteAddr:    &rmtAddr,
			PublicKeyAlgo: &pubAlgo,
			SignatureAlgo: &sigAlgo,
			KeyUsage:      keyUsageToString(cert.KeyUsage),
			ExtKeyUsages:  &keyUsages,
			PublicKey:     getPublicKeyType(cert),
			EncodedPEM:    encodedPemFromCert(cert),
			Signature:     sha1HexFromCertSignature(cert.Signature),
			Issued:        &cert.NotBefore,
			Expires:       &cert.NotAfter,
			DNSNames:      &dnsNames,
			Issuer:        &org,
			LastPollAt:    time.Now(),
			Latency:       &lt,
			Status:        checkCertificateStatus(cert.NotAfter),
		}
	}()

	select {
	case <-ctx.Done():
		stv = StatusUnResponsive
		str := ctx.Err().Error()
		return &TrackingDomainInfo{
			Error:      &str,
			LastPollAt: time.Now(),
			Status:     &stv,
		}, nil
	case result := <-resultch:
		return &result, nil
	}
}
