package ssl

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"net"
	"strings"
	"time"
)

func isVerificationError(err error) bool {
	// Check if the error is due to a verification failure
	_, isVerificationError := err.(*tls.CertificateVerificationError)
	return isVerificationError
}

func isConnectionError(err error) bool {
	// Check if the error is due to a connection refused
	_, isConnectionRefused := err.(*net.OpError)
	return isConnectionRefused
}

func extKeyUsageToString(usage x509.ExtKeyUsage) string {
	switch usage {
	case x509.ExtKeyUsageAny:
		return "Any"
	case x509.ExtKeyUsageServerAuth:
		return "Server Authentication"
	case x509.ExtKeyUsageClientAuth:
		return "Client Authentication"
	case x509.ExtKeyUsageCodeSigning:
		return "Code Signing"
	case x509.ExtKeyUsageEmailProtection:
		return "Email Protection"
	case x509.ExtKeyUsageIPSECEndSystem:
		return "IPsec End System"
	case x509.ExtKeyUsageIPSECTunnel:
		return "IPsec Tunnel"
	case x509.ExtKeyUsageIPSECUser:
		return "IPsec User"
	case x509.ExtKeyUsageTimeStamping:
		return "Timestamping"
	case x509.ExtKeyUsageOCSPSigning:
		return "OCSP Signing"
	default:
		return "Unknown"
	}
}

func keyUsageToString(usage x509.KeyUsage) *string {
	var usageStrings []string

	if usage&x509.KeyUsageDigitalSignature != 0 {
		usageStrings = append(usageStrings, "Digital Signature")
	}
	if usage&x509.KeyUsageContentCommitment != 0 {
		usageStrings = append(usageStrings, "Content Commitment")
	}
	if usage&x509.KeyUsageKeyEncipherment != 0 {
		usageStrings = append(usageStrings, "Key Encipherment")
	}
	if usage&x509.KeyUsageDataEncipherment != 0 {
		usageStrings = append(usageStrings, "Data Encipherment")
	}
	if usage&x509.KeyUsageKeyAgreement != 0 {
		usageStrings = append(usageStrings, "Key Agreement")
	}
	if usage&x509.KeyUsageCertSign != 0 {
		usageStrings = append(usageStrings, "Certificate Signing")
	}
	if usage&x509.KeyUsageCRLSign != 0 {
		usageStrings = append(usageStrings, "CRL Signing")
	}
	if usage&x509.KeyUsageEncipherOnly != 0 {
		usageStrings = append(usageStrings, "Encipher Only")
	}
	if usage&x509.KeyUsageDecipherOnly != 0 {
		usageStrings = append(usageStrings, "Decipher Only")
	}

	str := fmt.Sprintf("[%s]", joinStrings(usageStrings, ", "))
	return &str
}

func joinStrings(strings []string, sep string) string {
	if len(strings) == 0 {
		return ""
	}
	if len(strings) == 1 {
		return strings[0]
	}
	return strings[0] + sep + joinStrings(strings[1:], sep)
}

func getPublicKeyType(cert *x509.Certificate) *string {
	str := "Unknown"
	switch cert.PublicKey.(type) {
	case *rsa.PublicKey:
		str = "RSA"
		return &str
	case *ecdsa.PublicKey:
		str = "ECDSA"
		return &str
	default:
		return &str
	}
}

func encodedPemFromCert(cert *x509.Certificate) *string {
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	str := strings.TrimSpace(string(certPEM))
	return &str
}

func sha1HexFromCertSignature(signature []byte) *string {
	// Calculate SHA-1 hash of the certificate signature
	hash := sha1.Sum(signature)

	// Convert the hash to a hex string
	str := hex.EncodeToString(hash[:])
	return &str
}

func checkCertificateStatus(tm time.Time) *string {
	now := time.Now()

	if now.After(tm) {
		str := StatusExpired
		return &str
	}

	expirationThreshold := 30 * 24 * time.Hour // 30 days
	if now.Add(expirationThreshold).After(tm) {
		str := StatusExpires
		return &str
	}

	str := StatusHealthy
	return &str
}
