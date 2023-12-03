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
	LastAlertTime *time.Time // the last time of alert of domain's expiration or changes
}

type DomainTracking struct {
	ID         int64
	UserID     int64
	DomainName string
	TrackingDomainInfo
}

// PollDomain conducts a domain poll to gather information about the specified domain.
// It uses the provided context 'ctx' for handling timeouts and cancellations.
// Parameters:
//   - ctx: The context for handling deadlines and cancellations.
//   - domain: The domain string representing the target domain for polling.
//
// Returns:
//   - *TrackingDomainInfo: A pointer to the structure containing domain information.
//   - error: An error indicating any issues encountered during the polling process.
func PollDomain(ctx context.Context, domain string) (*TrackingDomainInfo, error) {
	var (
		// This is for the domain: "How much time will it take to respond?"
		start = time.Now()
		// For tunneling in goroutines: If an error is encountered within the waiting time, it returns the error. Otherwise, it sends information about the domain through the 'resultch' channel.
		resultch = make(chan TrackingDomainInfo)
		config   = &tls.Config{}
		stv      string
		// Latency in milliseconds
		lt int
	)

	go func() {
		// Establishes a secure TLS connection over TCP to the given domain on port 443.
		// Uses the 'tls.Dial' function, forming the connection address using the 'domain' variable
		// and applies the TLS configurations specified in the 'config' variable.
		conn, err := tls.Dial("tcp", fmt.Sprintf("%s:443", domain), config)
		if err != nil {
			// Capture error information and calculate latency since the start time
			stv = err.Error()
			lt = int(time.Since(start).Milliseconds())

			// Create TrackingDomainInfo structure with error, latency, and timestamp
			info := TrackingDomainInfo{
				LastPollAt: time.Now(),
				Error:      &stv,
				Latency:    &lt,
			}

			// Update status based on the type of error and send info through resultch
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

			// Exit the function after handling the error
			return
		}
		defer conn.Close()
		var (
			// Retrieve the TLS connection state and peer certificate from the connection 'conn'.
			state     = conn.ConnectionState()
			cert      = state.PeerCertificates[0] // Extract the first peer certificate
			keyUsages = ""
		)

		// Iterate through each Extended Key Usage field in the certificate.
		// Concatenate a string representation of each key usage to the 'keyUsages' variable.
		for _, usage := range cert.ExtKeyUsage {
			keyUsages += " " + extKeyUsageToString(usage)
		}

		// Collects information from the TLS certificate and connection.
		// Constructs a 'TrackingDomainInfo' structure and sends it through the 'resultch' channel.
		dnsNames := strings.Join(cert.DNSNames, ", ") // Join DNS names into a string
		org := cert.Issuer.Organization[0]            // Retrieve the organization of the certificate issuer
		lt = int(time.Since(start).Milliseconds())    // Calculate the latency
		pubAlgo := cert.PublicKeyAlgorithm.String()   // Get the public key algorithm
		sigAlgo := cert.SignatureAlgorithm.String()   // Get the signature algorithm
		rmtAddr := conn.RemoteAddr().String()         // Get the remote address of the connection
		// Create and send a 'TrackingDomainInfo' object through the channel
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

	// Select statement to handle context cancellation or receiving results from the result channel.
	select {
	case <-ctx.Done(): // Handle context cancellation or timeout
		stv = StatusUnResponsive
		str := ctx.Err().Error()
		return &TrackingDomainInfo{
			Error:      &str,
			LastPollAt: time.Now(),
			Status:     &stv,
		}, nil
	case result := <-resultch: // Receive result from the result channel
		return &result, nil
	}
}
