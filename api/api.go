package api

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/SaidovZohid/certalert.info/api/handlers"
	h "github.com/SaidovZohid/certalert.info/api/handlers"
	"github.com/SaidovZohid/certalert.info/config"
	"github.com/SaidovZohid/certalert.info/pkg/logger"
	"github.com/SaidovZohid/certalert.info/pkg/ssl"
	"github.com/SaidovZohid/certalert.info/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/django/v3"

	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type RoutetOptions struct {
	Cfg      *config.Config
	Log      logger.Logger
	Strg     storage.StorageI
	InMemory storage.InMemoryStorageI
}

func New(opt *RoutetOptions) *fiber.App {
	engine := django.New("views", ".html")
	engine.Reload(true)

	app := fiber.New(fiber.Config{
		EnableIPValidation:      true,
		Views:                   engine,
		EnableTrustedProxyCheck: true,
		PassLocalsToViews:       true,
		// WriteTimeout:            10 * time.Minute,
		ErrorHandler: HandleError,
	})

	// Redirect invalid API requests to the main URL
	app.Use(recover.New(recover.Config{EnableStackTrace: true}))
	//app.Use(func(c *fiber.Ctx) error {
	//	c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	//	return c.Next()
	//})
	app.Use(h.WithFlash)
	app.Use(favicon.New(favicon.Config{
		File: "./static/favicon.ico",
		URL:  "/favicon.ico",
	}))

	// handlers for html
	// engine.AddFunc("aside", func(id int) string {
	// 	log.Println("Something")
	// 	return fmt.Sprintf(aside, id)
	// })
	engine.AddFunc("expires", func(tm interface{}, name string) string {
		if tm == nil {
			return "unavailable"
		}

		timeNow := time.Now()
		expirationTime := tm.(*time.Time)

		if timeNow.Equal(*expirationTime) {
			return "expired"
		}

		daysLeft := expirationTime.Sub(timeNow).Hours() / 24
		if daysLeft <= 1 && daysLeft >= 0 {
			if name != "dashboard" {
				return "Expires in 1 day"
			}
			return "1 day"
		} else if daysLeft > 1 {
			if name != "dashboard" {
				return fmt.Sprintf("Expires in %.0f days", daysLeft)
			}
			return fmt.Sprintf("%.0f days", daysLeft)
		}

		return "expired"
	})
	engine.AddFunc("issuer", func(issuer interface{}) string {
		if issuer == nil {
			return "unavailable"
		}
		is := issuer.(*string)
		return *is
	})
	engine.AddFunc("domainStatus", func(status *string) string {
		switch *status {
		case ssl.StatusExpired:
			return fmt.Sprintf(`<td class="px-4 py-2 font-bold text-red-600 domain-status">%v</td>`, ssl.StatusExpired)
		case ssl.StatusHealthy:
			return fmt.Sprintf(`<td class="px-4 py-2 font-bold text-green-600 domain-status">%v</td>`, ssl.StatusHealthy)
		case ssl.StatusInvalid:
			return fmt.Sprintf(`<td class="px-4 py-2 font-bold text-yellow-600 domain-status">%v</td>`, ssl.StatusInvalid)
		case ssl.StatusOffline:
			return fmt.Sprintf(`<td class="px-4 py-2 font-bold text-gray-400 domain-status">%v</td>`, ssl.StatusOffline)
		case ssl.StatusUnResponsive:
			return fmt.Sprintf(`<td class="px-4 py-2 font-bold text-teal-600 domain-status">%v</td>`, ssl.StatusUnResponsive)
		case ssl.StatusExpires:
			return fmt.Sprintf(`<td class="px-4 py-2 font-bold text-orange-600 domain-status">%v</td>`, ssl.StatusExpires)
		}
		return `<td class="px-4 py-2 font-bold">unavailable</td>`
	})
	engine.AddFunc("domainStatusToString", func(domainName *string) string {
		if domainName == nil {
			return "unavailable"
		} else if *domainName == ssl.StatusHealthy {
			return ssl.StatusHealthy
		} else if *domainName == ssl.StatusExpires {
			return ssl.StatusExpires
		} else if *domainName == ssl.StatusExpired {
			return ssl.StatusExpired
		} else if *domainName == ssl.StatusInvalid {
			return ssl.StatusInvalid
		} else if *domainName == ssl.StatusOffline {
			return ssl.StatusOffline
		} else if *domainName == ssl.StatusUnResponsive {
			return ssl.StatusUnResponsive
		}
		return "unavailable"
	})
	engine.AddFunc("timeFormat", func(tm interface{}) string {
		if tm == nil {
			return "unavailable"
		}
		timeFormated := tm.(*time.Time).Format(time.RFC1123)
		return strings.Split(timeFormated, "+")[0]
	})

	engine.AddFunc("LastPollTimeFormat", func(tm time.Time, timeZone string) string {
		// Get the time zone location from the provided session time zone.
		loc, err := time.LoadLocation(timeZone)
		if err != nil {
			opt.Log.Error(err)
			return ""
		}

		// Convert the time to the user's local time zone.
		localTime := tm.In(loc)
		log.Println(localTime)
		log.Println(loc)

		// Format the local time.
		timeFormatted := localTime.Format(time.RFC1123)
		return strings.Split(timeFormatted, "+")[0]
	})

	engine.AddFunc("ipAddress", func(ip interface{}) string {
		if ip == nil {
			return "unavailable"
		}

		ipAddr := *ip.(*string)

		// Remove :443 from each IP address
		parts := strings.Split(ipAddr, ":")

		return parts[0]
	})

	engine.AddFunc("parseExtKeyUsage", func(extKeyUsages interface{}) string {
		if extKeyUsages == nil {
			return "unavailable"
		}

		return *extKeyUsages.(*string)
	})

	app.Static("/static", "./static")

	handlers := h.New(&h.HandlerV1Options{
		Cfg:                   opt.Cfg,
		Log:                   opt.Log,
		Strg:                  opt.Strg,
		InMemory:              opt.InMemory,
		Tokens:                make(map[string]handlers.TokenDataValidAndToken, 0),
		ForgotPasswordUserReq: make(map[string]string, 0),
	})

	app.Get("/", handlers.HandleGetLandingPage)
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("Verification link sent!")
	})
	app.Get("/error", func(c *fiber.Ctx) error {
		return errors.New("error")
	})

	// signup
	app.Get("/signup", handlers.SimpleAuthMiddleware, handlers.HandleGetSignUpPage)
	app.Post("/signup", handlers.SimpleAuthMiddleware, handlers.HandeSignupUser)
	app.Get("/signup/options", handlers.SimpleAuthMiddleware, handlers.HandleVerifyUserSignUp)

	// login
	app.Post("/login", handlers.SimpleAuthMiddleware, handlers.HandeLoginUser)
	app.Get("/login", handlers.SimpleAuthMiddleware, handlers.HandleGetLoginPage)

	// google login
	app.Get("/login/google", handlers.SimpleAuthMiddleware, handlers.HandleGoogleAuth)
	app.Get("/login/google/callback", handlers.SimpleAuthMiddleware, handlers.HandleGoogleCallback)

	app.Get("/logout", handlers.AuthMiddleware, handlers.HandleLogout)

	app.Get("/forgot-password", handlers.SimpleAuthMiddleware, handlers.HandleForgotPasswordPage)
	app.Get("/forgot-password/options", handlers.SimpleAuthMiddleware, handlers.HandleForgotPasswordVerify)
	app.Post("/forgot-password", handlers.SimpleAuthMiddleware, handlers.HandleForgotPassword)
	app.Post("/forgot-password/update", handlers.SimpleAuthMiddleware, handlers.HandleForgotPasswordUpdate)

	app.Get("/domains", handlers.AuthMiddleware, handlers.HandleDomainsPage)
	app.Post("/domains/add/new", handlers.AuthMiddleware, handlers.AddNewDomains)
	app.Get("/domains/add", handlers.AuthMiddleware, handlers.AddNewDomainsPage)
	app.Delete("/domains/stop", handlers.AuthMiddleware, handlers.HandleStopMonitoringDomains)
	app.Post("/domains/stop/:id", handlers.AuthMiddleware, handlers.HandleStopMonitoringDomain)
	app.Get("/domains/check", handlers.AuthMiddleware, handlers.HandleCheckDomains)
	app.Get("/domains/more/:id", handlers.AuthMiddleware, handlers.HandleDomainInfoShowPage)
	app.Get("/domains/pem/:id", handlers.AuthMiddleware, handlers.HandleShowEncodedPEM)

	app.Get("/account", handlers.AuthMiddleware, handlers.HandleAccountPage)
	app.Get("/account/delete", handlers.AuthMiddleware, handlers.HandleDeleteAccount)

	// change email
	app.Get("/account/change-email", handlers.AuthMiddleware, handlers.HandleChangeEmailPage)
	app.Post("/account/change-email", handlers.AuthMiddleware, handlers.HandleChangeEmail)
	app.Get("/account/change-email/options", handlers.HandleChangeEmailOptions)

	app.Use(func(c *fiber.Ctx) error {
		return c.Render("404/index", nil)
	})

	return app
}
