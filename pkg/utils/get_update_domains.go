package utils

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/SaidovZohid/certalert.info/config"
	"github.com/SaidovZohid/certalert.info/pkg/logger"
	"github.com/SaidovZohid/certalert.info/pkg/ssl"
	"github.com/SaidovZohid/certalert.info/storage"
	"github.com/SaidovZohid/certalert.info/storage/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type DomainNowAndPreviousInfo struct {
	DomainName string
	Current    *ssl.TrackingDomainInfo
	Prev       *ssl.TrackingDomainInfo
}

type UpdateDomainRegArgs struct {
	Strg storage.StorageI
	Log  *logger.Logger
	Cfg  *config.Config
	Bot  *tgbotapi.BotAPI
}

type UpdateDomainRegI interface {
	UpdateDomainInformationRegularly(ctx context.Context)
}

func NewUpdateReg(strg storage.StorageI, log logger.Logger, cfg *config.Config, bot *tgbotapi.BotAPI) UpdateDomainRegI {
	return &UpdateDomainRegArgs{
		Strg: strg,
		Log:  &log,
		Cfg:  cfg,
		Bot:  bot,
	}
}

// UpdateDomainInformationRegularly fetches information about domains at regular intervals
// of x minutes, pausing in between iterations.
func (args *UpdateDomainRegArgs) UpdateDomainInformationRegularly(ctx context.Context) {
	//
	ticker := time.NewTicker(args.Cfg.PullUpdateDomainInterval)
	done := make(chan struct{})

	if err := args.poll(ctx); err != nil {
		args.Log.Errorf("Failed to pull data from certificate transparency logs for all domains: %s", err)
		return
	}

	args.Log.Info("Successfully pulled info and updated for the first time!")

	// Run the domain information update loop
	for {
		select {
		case <-ticker.C:
			args.Log.Info("Domain information update started")

			if err := args.poll(ctx); err != nil {
				args.Log.Errorf("Failed to pull data from certificate transparency logs for all domains: %s", err)
				continue
			}

			args.Log.Info("Successfully pulled info and updated!")
		case <-done:
			args.Log.Info("UpdateDomainInformationRegularly has been canceled!")
			return // If the context is canceled, exit the function
		}
	}
}

func (args *UpdateDomainRegArgs) poll(ctx context.Context) error {
	domains, err := args.Strg.Domain().GetListofDomainsThatExists(ctx)
	if err != nil {
		args.Log.Errorf("Failed to get list of domains from storage: %s", err)
		return err
	}

	var (
		workers = make(chan struct{}, 50)
		wg      = sync.WaitGroup{}
		results = make(chan DomainNowAndPreviousInfo, len(domains))
	)
	defer close(workers)
	args.Log.Info("Domains -> ", len(domains))
	for _, domain := range domains {
		wg.Add(1)
		go func(domain *ssl.DomainTracking) {
			ctxPoll, cancel := context.WithTimeout(context.Background(), time.Second*10)
			workers <- struct{}{}
			defer func() {
				<-workers
				wg.Done()
				cancel()
			}()

			info, err := ssl.PollDomain(ctxPoll, domain.DomainName)
			if err != nil {
				args.Log.Error(err)
				return
			}

			err = args.Strg.Domain().UpdateAllTheSameDomainsInfo(ctx, &ssl.DomainTracking{
				DomainName:         domain.DomainName,
				TrackingDomainInfo: *info,
			})
			if err != nil {
				args.Log.Error(err)
				return
			}

			results <- DomainNowAndPreviousInfo{
				DomainName: domain.DomainName,
				Prev:       &domain.TrackingDomainInfo,
				Current:    info,
			}
		}(domain)
	}

	wg.Wait()
	close(results)

	return args.filterDomainsOwnersNotif(ctx, results)
}

func (args *UpdateDomainRegArgs) filterDomainsOwnersNotif(ctx context.Context, results chan DomainNowAndPreviousInfo) error {
	args.Log.Info("In process of sending notification to the users ")

	// TODO:
	// * for loop the domain and give the domain to strg and get the users info in Query.
	// * run loop inside the results loop.
	// * in for loop one by one get the user and check which things are turned on in user setting for notification.
	// * if email is turned on, send notification throw email
	// * if telegram is turned on, send notification throw telegram bot. if it turned on, try to get the user telegram id that is linked to the user account, and send the chat id within bot.
	for v := range results {
		users, err := args.Strg.Domain().GetListofUsersThatDomainExists(ctx, v.DomainName)
		if err != nil {
			args.Log.Errorf("error getting list of user that has this domain %s", err)
			continue
		}
		for _, userId := range users {
			domain, err := args.Strg.Domain().GetDomainWithUserIDAndDomainName(ctx, &ssl.DomainTracking{
				DomainName: v.DomainName,
				UserID:     userId,
			})
			if err != nil {
				args.Log.Errorf("error getting domain with user id and domain name %s", err)
				continue
			}
			// If last alert time is one day after, send notification about expiration. Otherwise just skip it.
			if isLastAlertTimeOneDayAgo(*domain.LastAlertTime) {
				user, err := args.Strg.User().GetUserByID(ctx, userId)
				if err != nil {
					args.Log.Errorf("error getting user by id %d", err)
					continue
				}
				notification, err := args.Strg.Notifications().GetNotificationRowByUserID(ctx, userId)
				if err != nil {
					args.Log.Errorf("error getting notification row by userid %d", err)
					continue
				}
				err = args.notifyUser(ctx, user, notification, &v)
				if err != nil {
					args.Log.Errorf("error notifying user %s", err)
					continue
				}
			}
		}
	}
	return nil
}

func isLastAlertTimeOneDayAgo(lastAlertTime time.Time) bool {
	// Calculate the duration between current time and last_alert_time
	duration := time.Since(lastAlertTime)

	// if the duration is greater than or equal to 24 hours; return true; return false
	return duration >= 24*time.Hour
}

var changeAlertStr = "change_alert"
var expiryAlertStr = "expiry_alert"

func (args *UpdateDomainRegArgs) notifyUser(ctx context.Context, user *models.User, notification *models.Notification, domainPrInfo *DomainNowAndPreviousInfo) error {
	expiryAlert, changeAlert := checkExpiryAndChangeSSLOfDomain(user, domainPrInfo, notification)
	// TODO:
	// * check the expiry or change alert true or false and write the logic of sending of notification code!
	var isNotified bool
	if notification.ExpiryAlerts && expiryAlert {
		args.Log.Info("Expiration Notify ", domainPrInfo.DomainName)
		if err := args.sendNotificationChangeOrExpire(&expiryAlertStr, user, domainPrInfo, notification); err != nil {
			return err
		}
		isNotified = true
	} else if notification.ChangeAlert && changeAlert {
		args.Log.Info("Change Notify ", domainPrInfo.DomainName)
		if err := args.sendNotificationChangeOrExpire(&changeAlertStr, user, domainPrInfo, notification); err != nil {
			return err
		}
		isNotified = true
	}
	if isNotified {
		return args.Strg.Domain().UpdateTheLastAlertTime(ctx, user.ID, domainPrInfo.DomainName)
	}

	return nil
}

func checkExpiryAndChangeSSLOfDomain(user *models.User, domainPrInfo *DomainNowAndPreviousInfo, notification *models.Notification) (expiryAlert bool, changeAlert bool) {
	if domainPrInfo.Current.Expires == nil {
		return domainPrInfo.Current.Expires != domainPrInfo.Prev.Expires, true
	}
	expirationDate := domainPrInfo.Current.Expires.AddDate(0, 0, -notification.Before)
	currentDate := time.Now()

	expiryAlert = currentDate.After(expirationDate)

	if domainPrInfo.Current.RemoteAddr != nil {
		changeAlert = *domainPrInfo.Prev.RemoteAddr != *domainPrInfo.Current.RemoteAddr || *domainPrInfo.Prev.Issuer != *domainPrInfo.Current.Issuer ||
			*domainPrInfo.Prev.PublicKey != *domainPrInfo.Current.PublicKey ||
			*domainPrInfo.Prev.DNSNames != *domainPrInfo.Current.DNSNames ||
			*domainPrInfo.Prev.KeyUsage != *domainPrInfo.Current.KeyUsage ||
			*domainPrInfo.Prev.ExtKeyUsages != *domainPrInfo.Current.ExtKeyUsages ||
			*domainPrInfo.Prev.Status != *domainPrInfo.Current.Status
	}

	return expiryAlert, changeAlert
}

// tp = {change_alert or expiry_alert}
func (args *UpdateDomainRegArgs) sendNotificationChangeOrExpire(tp *string, user *models.User, domainPrInfo *DomainNowAndPreviousInfo, notification *models.Notification) error {
	if tp == nil {
		return errors.New("nil notification type")
	}
	var err error
	switch *tp {
	case expiryAlertStr:
		if notification.EmailAlert {
			err = args.sendNotificationToUserByEmail(tp, user, domainPrInfo, notification)
			if err != nil {
				return err
			}
		}
		if notification.TelegramAlert {
			err = args.sendNotificationToUserByTelegram(tp, user, domainPrInfo, notification)
			if err != nil {
				return err
			}
		} else {
			err = errors.New("no alert method is set")
		}
	case changeAlertStr:

	default:
		err = fmt.Errorf("unknown type %v", tp)
	}

	return err
}

// tp = {change_alert or expiry_alert}
// Email Notification
func (args *UpdateDomainRegArgs) sendNotificationToUserByEmail(tp *string, user *models.User, domainPrInfo *DomainNowAndPreviousInfo, notification *models.Notification) error {
	if tp == nil {
		return errors.New("nil notification type")
	}
	return nil
}

// tp = {change_alert or expiry_alert}
// Telegram Notification
func (args *UpdateDomainRegArgs) sendNotificationToUserByTelegram(tp *string, user *models.User, domainPrInfo *DomainNowAndPreviousInfo, notification *models.Notification) error {
	if tp == nil {
		return errors.New("nil notification type")
	}
	userTg, err := args.Strg.Integrations().GetFromTelegramByUserID(context.Background(), user.ID)
	if err != nil {
		return err
	}
	if userTg.ChatID == 0 {
		return fmt.Errorf("no chat id for telegram notification to user id %v", user.ID)
	}
	var msg string
	if userTg.Lang == "uz" {
		msg = "Assalomu Alaykum 👋️️️️,\n\nSiz kuzatayotgan domen, " + domainPrInfo.DomainName + ", "
	} else if userTg.Lang == "ru" {
		msg = "Здравствуйте 👋️️️️,\n\nВаш отслеживаемый домен, " + domainPrInfo.DomainName + ", "
	} else if userTg.Lang == "eng" {
		msg = "Hello 👋️️️️,\n\nYour tracked domain, " + domainPrInfo.DomainName + ", "
	} else {
		return fmt.Errorf("unsupported language code %s", userTg.Lang)
	}
	switch *tp {
	case expiryAlertStr:
		lft := daysUntilExpiration(*domainPrInfo.Current.Expires)
		if userTg.Lang == "uz" {
			msg += fmt.Sprintf("yaqinlashib kelayotgan SSL muddati bor. Faqat [%v] kun qoldi. Zudlik bilan harakat qiling-tafsilotlarni tekshiring [%v].", lft, args.Cfg.BaseUrl)
		} else if userTg.Lang == "ru" {
			msg += fmt.Sprintf("истекает срок действия SSL. Осталось всего [%v] дней. Действуйте незамедлительно - проверьте подробности на [%v].", lft, args.Cfg.BaseUrl)
		} else if userTg.Lang == "eng" {
			msg += fmt.Sprintf("has an upcoming SSL expiration. Only [%v] days left. Act promptly - check details at [%v].", lft, args.Cfg.BaseUrl)
		} else {
			return fmt.Errorf("unsupported language code %s", userTg.Lang)
		}
	case changeAlertStr:
		args.Log.Info("sending change alert notification")
	default:
		return errors.New("unknown type " + *tp)
	}

	message := tgbotapi.NewMessage(userTg.ChatID, msg)
	if _, err := args.Bot.Send(message); err != nil {
		return err
	}

	return nil
}

func daysUntilExpiration(expirationTime time.Time) int {
	// Calculate the duration until expiration
	duration := time.Until(expirationTime)

	// Convert the duration to days
	daysLeft := int(duration.Hours() / 24)

	fmt.Println(daysLeft)

	return daysLeft
}
