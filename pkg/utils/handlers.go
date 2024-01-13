package utils

import (
	"fmt"
	"github.com/SaidovZohid/certalert.info/pkg/ssl"
	"log"
	"time"

	"github.com/SaidovZohid/certalert.info/storage/models"
)

var changeAlertStr = "change_alert"
var expiryAlertStr = "expiry_alert"

// returns is lastAlertTime is one day bigger or not. If bigger one day returns true, otherwise false
func isLastAlertTimeOneDayAgo(lastAlertTime time.Time) bool {
	// Calculate the duration between current time and last_alert_time
	duration := time.Since(lastAlertTime)

	// if the duration is greater than or equal to 24 hours; return true; return false
	return duration >= 24*time.Hour
}

// checks the expiration and change on ssl. If expiration has returns true, otherwise false. If change has on domain's ssl returns true, otherwise false
func checkExpiryAndChangeSSLOfDomain(domainPrInfo *DomainNowAndPreviousInfo, notification *models.Notification) (expiryAlert, changeAlert bool) {
	log.Println(*domainPrInfo.Current)

	// Check if the current expiration date is nil
	if domainPrInfo.Current.Expires == nil {
		return domainPrInfo.Current.Expires != domainPrInfo.Prev.Expires, true
	}

	// Calculate the expiration date with the notification period
	expirationDate := domainPrInfo.Current.Expires.AddDate(0, 0, -notification.Before)

	// Get the current date
	currentDate := time.Now()

	// Check if the certificate has expired
	expiryAlert = currentDate.After(expirationDate)

	// Check for changes in certificate details if RemoteAddr is not nil
	if domainPrInfo.Current.RemoteAddr != nil {
		changeAlert = hasCertificateDetailsChanged(domainPrInfo.Prev, domainPrInfo.Current)
	}

	return expiryAlert, changeAlert
}

// hasCertificateDetailsChanged checks for changes in certificate details
func hasCertificateDetailsChanged(prev, current *ssl.TrackingDomainInfo) bool {
	return *prev.RemoteAddr != *current.RemoteAddr ||
		*prev.Issuer != *current.Issuer ||
		*prev.PublicKey != *current.PublicKey ||
		*prev.DNSNames != *current.DNSNames ||
		*prev.KeyUsage != *current.KeyUsage ||
		*prev.ExtKeyUsages != *current.ExtKeyUsages ||
		*prev.Status != *current.Status
}

// returns how many days left to ssl of domain expiration
func daysUntilExpiration(expirationTime time.Time) int {
	// Calculate the duration until expiration
	duration := time.Until(expirationTime)

	// Convert the duration to days (rounded to the nearest integer)
	daysLeft := int(duration.Hours()/24 + 0.5)

	fmt.Println(daysLeft)

	return daysLeft
}
