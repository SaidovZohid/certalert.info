package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/SaidovZohid/certalert.info/config"
	"github.com/SaidovZohid/certalert.info/pkg/logger"
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
}

type HandlerV1Options struct {
	Cfg      *config.Config
	Log      logger.Logger
	Strg     storage.StorageI
	InMemory storage.InMemoryStorageI
}

func New(options *HandlerV1Options) *handlerV1 {
	return &handlerV1{
		cfg:      options.Cfg,
		log:      options.Log,
		strg:     options.Strg,
		inMemory: options.InMemory,
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
	log.Println(userinfo)

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
