package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/SaidovZohid/certalert.info/pkg/logger"
	"github.com/SaidovZohid/certalert.info/pkg/ssl"
	"github.com/SaidovZohid/certalert.info/storage"
	"github.com/SaidovZohid/certalert.info/storage/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func lookThroughDomainInfoAndExpiryAndChange(userID int64, domains []*DomainNowAndPreviousInfo, strg storage.StorageI, log logger.Logger, bot *tgbotapi.BotAPI) {
	log.Info("System is sending notification!")
	notificationUser, err := strg.Notifications().GetNotificationRowByUserID(context.Background(), userID)
	if err != nil {

		log.Error(err)
		return
	}
	log.Info("I am here after GetNotificationRowByUserID func")
	tr := isMoreThanOneDayAhead(notificationUser.LastAlertTime)
	log.Info(tr)
	if tr {
		domainsToNotifyChange := make([]*ssl.DomainTracking, 0)
		domainsToNotifyExpiry := make([]*ssl.DomainTracking, 0)
		for _, domain := range domains {
			if domain.Current.Error != nil {
				// Trigger alert or log the changes
				log.Warnf("Changes detected for domain %s", domain.DomainName)
				if notificationUser.ChangeAlert && hasChanges(domain.Prev, domain.Current) {
					domainsToNotifyChange = append(domainsToNotifyChange, &ssl.DomainTracking{
						DomainName:         domain.DomainName,
						UserID:             userID,
						TrackingDomainInfo: *domain.Current,
					})
					continue
				}
				if notificationUser.ExpiryAlerts {
					now := time.Now()
					if now.After(*domain.Current.Expires) {
						domainsToNotifyExpiry = append(domainsToNotifyExpiry, &ssl.DomainTracking{
							DomainName:         domain.DomainName,
							UserID:             userID,
							TrackingDomainInfo: *domain.Current,
						})
						continue
					} else if shouldSendExpirationAlert(*domain.Current.Expires, notificationUser.Before) {
						domainsToNotifyExpiry = append(domainsToNotifyExpiry, &ssl.DomainTracking{
							DomainName:         domain.DomainName,
							UserID:             userID,
							TrackingDomainInfo: *domain.Current,
						})
						continue
					}
				}
			} else {
				if notificationUser.ChangeAlert && hasChanges(domain.Prev, domain.Current) {
					domainsToNotifyChange = append(domainsToNotifyChange, &ssl.DomainTracking{
						DomainName:         domain.DomainName,
						UserID:             userID,
						TrackingDomainInfo: *domain.Current,
					})
					continue
				}
			}
		}

		sendNotificationToUser(userID, domainsToNotifyChange, domainsToNotifyExpiry, notificationUser, bot, strg, log)
	}
}

func sendNotificationToUser(userID int64, domainsToNotifyChange []*ssl.DomainTracking, domainsToNotifyExpiry []*ssl.DomainTracking, notificationUser *models.Notification, bot *tgbotapi.BotAPI, strg storage.StorageI, log logger.Logger) {
	if notificationUser.TelegramAlert {
		tgUser, err := strg.Integrations().GetFromTelegramByUserID(context.Background(), userID)
		if err != nil {
			log.Errorf("could not get telegram integration of user id - %v", userID)
		}
		telegramMessage := ""
		switch tgUser.Lang {
		case "en":
			telegramMessage += messageTextForNotificationInEnglishLang(domainsToNotifyChange, domainsToNotifyExpiry)
		case "uz":
		case "ru":
		}
		msg := tgbotapi.NewMessage(tgUser.ChatID, telegramMessage)
		msg.ParseMode = tgbotapi.ModeMarkdown

		if _, err := bot.Send(msg); err != nil {
			log.Errorf("userID - %v --- err - %v\n", userID, err)
		}
		// if domainsToNotifyExpiry != nil {
		// 	switch tgUser.Lang {
		// 	case "en":
		// 		telegramMessage = "*Expiring Soon and Expired:*\n\n"
		// 		for i, v := range domainsToNotifyExpiry {
		// 			telegramMessage += fmt.Sprintf("%v. *%v*\n\tExpiration Date: %v\n\n", i, v.DomainName, v.Expires.Format(time.RFC1123))
		// 		}
		// 		// msg.ParseMode = tgbotapi.ModeMarkdown
		// 	case "uz":
		// 		telegramMessage += fmt.Sprintf("")
		// 	case "ru":
		// 		telegramMessage += fmt.Sprintf("")
		// 	}
		// }
		// if domainsToNotifyChange != nil {
		// 	switch tgUser.Lang {
		// 	case "en":
		// 		telegramMessage = "*Changed:*\n\n"
		// 		for i, v := range domainsToNotifyExpiry {
		// 			telegramMessage += fmt.Sprintf("%v. *%v*\n\tExpiration Date: %v\n\n", i, v.DomainName, v.Expires.Format(time.RFC1123))
		// 		}
		// 		// msg.ParseMode = tgbotapi.ModeMarkdown
		// 	case "uz":
		// 		telegramMessage += fmt.Sprintf("")
		// 	case "ru":
		// 		telegramMessage += fmt.Sprintf("")
		// 	}
		// }
		// TODO: Write the logic of sending message alert to user via telegram.
	}
	if notificationUser.DiscordAlert {
		// TODO: Write the logic of sending message alert to user via discord.
	}
	if notificationUser.SlackAlert {
		// TODO: Write the logic of sending message alert to user via slack.
	}
	if notificationUser.MicrosoftTeamsAlert {
		// TODO: Write the logic of sending message alert to user via microsoft teams.
	}
	if notificationUser.EmailAlert {
		// TODO: Write the logic of sending email alert to user.
	}
}

func messageTextForNotificationInEnglishLang(domainsToNotifyChange []*ssl.DomainTracking, domainsToNotifyExpiry []*ssl.DomainTracking) string {
	txt := "Dear *user*,\n\n*Attention!* Your domain(s) require immediate action:\n\n"
	if domainsToNotifyExpiry != nil {
		txt += "*Expiring Soon and Expired:*\n\n"
		for i, v := range domainsToNotifyExpiry {
			txt += fmt.Sprintf("%v. *%v*\n\tExpiration Date: %v\n\n", i, v.DomainName, v.Expires.Format(time.RFC1123))
		}
	}
	if domainsToNotifyChange != nil {
		txt += "*Changed Information:*\n\n"
		for i, v := range domainsToNotifyChange {
			txt += fmt.Sprintf("%v. *%v*\n", i, v.DomainName)
		}
	}

	return txt
}

func shouldSendExpirationAlert(expirationTime time.Time, daysBefore int) bool {
	// Calculate the duration before expiration
	alertThreshold := time.Duration(daysBefore) * 24 * time.Hour // Convert days to hours

	// Calculate the time before which the alert should be sent
	alertTime := expirationTime.Add(-alertThreshold)

	// Compare with the current time
	return time.Now().After(alertTime)
}

func isMoreThanOneDayAhead(checkTime time.Time) bool {
	// Calculate the difference between the given time and the current time
	timeDifference := time.Now().Sub(checkTime)
	fmt.Println(timeDifference)
	// Check if the difference is more than 1 day
	return timeDifference > 24*time.Hour
}

func hasChanges(prev, current *ssl.TrackingDomainInfo) bool {
	// Compare each field and return true if any field is different
	return *prev.RemoteAddr != *current.RemoteAddr ||
		*prev.Issuer != *current.Issuer ||
		*prev.SignatureAlgo != *current.SignatureAlgo ||
		*prev.PublicKeyAlgo != *current.PublicKeyAlgo ||
		*prev.EncodedPEM != *current.EncodedPEM ||
		*prev.PublicKey != *current.PublicKey ||
		*prev.Signature != *current.Signature ||
		*prev.DNSNames != *current.DNSNames ||
		*prev.KeyUsage != *current.KeyUsage ||
		*prev.ExtKeyUsages != *current.ExtKeyUsages ||
		!prev.Issued.Equal(*current.Issued) ||
		!prev.Expires.Equal(*current.Expires) ||
		*prev.Status != *current.Status
}
