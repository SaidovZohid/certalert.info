package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/SaidovZohid/certalert.info/config"
	"github.com/SaidovZohid/certalert.info/pkg/logger"
	"github.com/SaidovZohid/certalert.info/pkg/ssl"
	"github.com/SaidovZohid/certalert.info/pkg/utils"
	"github.com/SaidovZohid/certalert.info/storage"
	"github.com/SaidovZohid/certalert.info/storage/models"
	"github.com/gofiber/fiber/v2"
	"github.com/ipinfo/go/v2/ipinfo"
	"github.com/mssola/useragent"
)

type handlerV1 struct {
	cfg      *config.Config
	log      logger.Logger
	strg     storage.StorageI
	inMemory storage.InMemoryStorageI
	tokens   map[string]string
}

type HandlerV1Options struct {
	Cfg      *config.Config
	Log      logger.Logger
	Strg     storage.StorageI
	InMemory storage.InMemoryStorageI
	Tokens   map[string]string
}

func New(options *HandlerV1Options) *handlerV1 {
	return &handlerV1{
		cfg:      options.Cfg,
		log:      options.Log,
		strg:     options.Strg,
		inMemory: options.InMemory,
		tokens:   options.Tokens,
	}
}

type User struct {
	FirstName string `json:"given_name"`
	LastName  string `json:"family_name"`
	Email     string `json:"email"`
}

func (h *handlerV1) getUserInfoFromGoogle(token string) (*User, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token)
	if err != nil {
		return nil, err
	}

	userdata, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	userinfo := make(map[string]interface{}, 0)
	err = json.Unmarshal(userdata, &userinfo)
	if err != nil {
		return nil, err
	}

	var data User
	data.Email = userinfo["email"].(string)
	data.FirstName = userinfo["given_name"].(string)
	data.LastName = userinfo["family_name"].(string)

	return &data, nil
}

func (h *handlerV1) SetCookie(c *fiber.Ctx, name string, val string, expires time.Time) {
	if h.cfg.BaseUrl == "http://localhost:3000" {
		c.Cookie(&fiber.Cookie{
			Name:     h.cfg.AuthCookieNameCertAlert,
			Value:    val,
			Path:     "/",
			Expires:  expires,
			HTTPOnly: true,
			Secure:   true,
			SameSite: fiber.CookieSameSiteLaxMode,
		})
	} else {
		c.Cookie(&fiber.Cookie{
			Name:     h.cfg.AuthCookieNameCertAlert,
			Domain:   h.cfg.BaseUrl[8:],
			Value:    val,
			Path:     "/",
			Expires:  expires,
			HTTPOnly: true,
			Secure:   true,
			SameSite: fiber.CookieSameSiteLaxMode,
		})
	}
}

type LocationInfo struct {
	IP       string `json:"ip"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Timezone string `json:"timezone"`
}

func GetLocation(ipaddress string, cfg *config.Config) (*LocationInfo, error) {
	// params: httpClient, cache, token. `http.DefaultClient` and no cache will be used in case of `nil`.
	client := ipinfo.NewClient(nil, nil, cfg.LocationInfoKey)

	info, err := client.GetIPInfo(net.ParseIP(ipaddress))
	if err != nil {
		return nil, err
	}

	locationInfo := LocationInfo{
		City:     info.City,
		Country:  info.Country,
		Region:   info.Region,
		Timezone: info.Timezone,
		IP:       info.IP.String(),
	}

	return &locationInfo, nil
}

func handleLoginDependencies(c *fiber.Ctx, h *handlerV1, id int64, data *User) error {
	accessToken, payload, err := utils.CreateToken(h.cfg, &utils.TokenParams{
		UserID:   id,
		Duration: time.Hour * 24, // token is valid for 1 days
		Email:    data.Email,
	})
	if err != nil {
		return errors.New("failed to create jwt token, try again")
	}

	ipAddress := c.Get("X-Forwarded-For")

	locationInfo, err := GetLocation(ipAddress, h.cfg)
	if err != nil {
		log.Println("Failed to get user info: ", err)
	}

	// Parse the User-Agent string using the user_agent library
	ua := useragent.New(c.Get("User-Agent"))

	// Get device and OS details
	device := "Desktop"
	if ua.Mobile() {
		device = "Mobile"
	}
	os := ua.OS()
	// Get browser details
	browserName, browserVersion := ua.Browser()

	timezone := time.FixedZone("GMT+5", 5*60*60) // 5 hours ahead of UTC

	s := models.Session{
		UserId:      id,
		AccessToken: accessToken,
		IpAddress:   ipAddress,
		UserAgent:   fmt.Sprintf("%v, %v, %v-%v", device, os, browserName, browserVersion),
		ExpiresAt:   payload.ExpiredAt,
		LastLogin:   time.Now().In(timezone).Format(time.RFC1123),
	}

	if locationInfo != nil {
		s.City = locationInfo.City
		s.Country = locationInfo.Country
		s.Region = locationInfo.Region
		s.Timezone = locationInfo.Timezone
		s.IpAddress = locationInfo.IP
	}

	sessionID, err := h.strg.Session().Session(context.Background(), &s)
	if err != nil {
		return err
	}

	// Set cookie for 1 day
	h.SetCookie(c, h.cfg.AuthCookieNameCertAlert, sessionID.ID.String(), time.Now().Add(time.Hour*24))

	return nil
}

func getCurrentTimeInTimeZone(timezone string) (time.Time, error) {
	// Get the UTC offset for the timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}

	_, offsetSeconds := time.Now().In(loc).Zone()

	// Create a fixed timezone with the offset
	fixedTimezone := time.FixedZone(timezone, offsetSeconds)

	// Get the current time in the specified timezone
	currentTime := time.Now().In(fixedTimezone)

	return currentTime, nil
}

type TrackDomainAdd struct {
	Domains []string
	UserID  int64
	Log     *logger.Logger
	Strg    storage.StorageI
}

func TrackDomainsAdded(t *TrackDomainAdd) error {
	var (
		workers = make(chan struct{}, 15)
		wg      = sync.WaitGroup{}
	)
	for _, domain := range t.Domains {
		hasDomainInDB, err := t.Strg.Domain().GetDomainWithUserIDAndDomainName(context.Background(), &ssl.DomainTracking{
			UserID:     t.UserID,
			DomainName: domain,
		})
		if (err != nil && !errors.Is(err, sql.ErrNoRows)) || hasDomainInDB != nil {
			t.Log.Error(err)
			continue
		}
		wg.Add(1)
		go func(domain string) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			workers <- struct{}{}
			defer func() {
				<-workers
				wg.Done()
				cancel()
			}()

			info, err := ssl.PollDomain(ctx, domain)
			if err != nil {
				t.Log.Error(err)
				return
			}

			domainInfo, err := t.Strg.Domain().CreateTrackingDomain(context.Background(), &ssl.DomainTracking{
				UserID:             t.UserID,
				DomainName:         domain,
				TrackingDomainInfo: *info,
			})
			if err != nil {
				t.Log.Error(err)
				return
			}

			_ = domainInfo
		}(domain)
	}

	wg.Wait()

	return nil
}

func (h *handlerV1) CheckExistingDomains(userID int64, domains []string) error {
	var (
		workers           = make(chan struct{}, 15)
		wg                = sync.WaitGroup{}
		mutex             = sync.Mutex{}
		domainInfoUpdates []ssl.DomainTracking // Slice to store domain info updates

	)

	responseChannel := make(chan ssl.DomainTracking, len(domains))
	for _, domain := range domains {
		hasDomainInDB, err := h.strg.Domain().GetDomainWithUserIDAndDomainName(context.Background(), &ssl.DomainTracking{
			UserID:     userID,
			DomainName: domain,
		})
		if (err != nil && errors.Is(err, sql.ErrNoRows)) || hasDomainInDB == nil {
			h.log.Error(err)
			continue
		}
		wg.Add(1)
		go func(domain string, id int64) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			workers <- struct{}{}
			defer func() {
				<-workers
				wg.Done()
				cancel()
			}()

			info, err := ssl.PollDomain(ctx, domain)
			if err != nil {
				h.log.Error(err)
				return
			}

			responseChannel <- ssl.DomainTracking{
				UserID:             userID,
				ID:                 id,
				TrackingDomainInfo: *info,
			}

			_ = info
		}(domain, hasDomainInDB.ID)
	}

	wg.Wait()

	close(responseChannel) // Close the response channel

	// Collect responses and append to the domainInfoUpdates slice
	for domainInfo := range responseChannel {
		domainInfoUpdates = append(domainInfoUpdates, domainInfo)
	}

	// Update domain info in a separate loop to avoid locking while updating
	for _, info := range domainInfoUpdates {
		mutex.Lock()
		err := h.strg.Domain().UpdateExistingDomainInfo(context.Background(), &info)
		mutex.Unlock()
		if err != nil {
			h.log.Error(err)
		}
	}

	err := h.strg.User().UpdateUserLastPoll(context.Background(), userID)
	if err != nil {
		return err
	}

	return nil
}

var htmlCode = `
<div
  id="alert-1"
  class="flex items-center p-4 mb-4 mt-4 text-gray-900 rounded-lg bg-blue-50 dark:bg-gray-800 dark:text-white"
  role="alert"
>
  <svg
    class="flex-shrink-0 w-4 h-4"
    aria-hidden="true"
    xmlns="http://www.w3.org/2000/svg"
    fill="currentColor"
    viewBox="0 0 20 20"
  >
    <path
      d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5ZM9.5 4a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3ZM12 15H8a1 1 0 0 1 0-2h1v-3H8a1 1 0 0 1 0-2h2a1 1 0 0 1 1 1v4h1a1 1 0 0 1 0 2Z"
    />
  </svg>
  <span class="sr-only">Info</span>
  <div class="ml-3 text-sm font-medium">%v</div>
</div>
`