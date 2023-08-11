package api

import (
	"errors"
	"fmt"
	"strings"
	"time"

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
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		return c.Next()
	})
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
	engine.AddFunc("expires", func(tm interface{}) string {
		if tm == nil {
			return "Unavailable"
		}

		timeNow := time.Now()
		expirationTime := tm.(*time.Time)

		if timeNow.Equal(*expirationTime) {
			return "Unavailable"
		}

		daysLeft := expirationTime.Sub(timeNow).Hours() / 24
		if daysLeft == 1 {
			return "1 day"
		} else if daysLeft > 1 {
			return fmt.Sprintf("%.0f days", daysLeft)
		}

		return "Expired"
	})
	engine.AddFunc("issuer", func(issuer interface{}) string {
		if issuer == nil {
			return "Unavailable"
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
			return fmt.Sprintf(`<td class="px-4 py-2 font-bold text-teal-600 domain-status">%v</td>`, ssl.StatusOffline)
		case ssl.StatusUnResponsive:
			return fmt.Sprintf(`<td class="px-4 py-2 font-bold text-gray-400 domain-status">%v</td>`, ssl.StatusUnResponsive)
		case ssl.StatusExpires:
			return fmt.Sprintf(`<td class="px-4 py-2 font-bold text-orange-600 domain-status">%v</td>`, ssl.StatusExpires)
		}
		return `<td class="px-4 py-2 font-bold">Unavailable</td>`
	})
	// engine.AddFunc("noDomains", func() string {
	// 	return `<h1 class="text-black">No domains to track</h1>`
	// })

	engine.AddFunc("ipAddress", func(ip interface{}) string {
		if ip == nil {
			return "Unavailable"
		}

		ipAddr := *ip.(*string)

		// Remove :443 from each IP address
		parts := strings.Split(ipAddr, ":")

		return parts[0]
	})

	app.Static("/static", "./static")

	handlers := h.New(&h.HandlerV1Options{
		Cfg:      opt.Cfg,
		Log:      opt.Log,
		Strg:     opt.Strg,
		InMemory: opt.InMemory,
		Tokens:   make(map[string]string, 0),
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
	app.Post("/signup/user", handlers.SimpleAuthMiddleware, handlers.HandeSignupUser)
	app.Get("/signup/options", handlers.SimpleAuthMiddleware, handlers.HandleVerifyUserSignUp)

	// login
	app.Post("/login", handlers.SimpleAuthMiddleware, handlers.HandeLoginUser)
	app.Get("/login", handlers.SimpleAuthMiddleware, handlers.HandleGetLoginPage)

	// google login
	app.Get("/login/google", handlers.SimpleAuthMiddleware, handlers.HandleGoogleAuth)
	app.Get("/login/google/callback", handlers.SimpleAuthMiddleware, handlers.HandleGoogleCallback)

	app.Get("/logout", handlers.AuthMiddleware, handlers.HandleLogout)

	app.Get("/forgot-password", handlers.SimpleAuthMiddleware, handlers.HandleForgotPasswordPage)
	// TODO: write the handler of reset link checker
	app.Get("/forgot-password/options", handlers.SimpleAuthMiddleware, handlers.HandleForgotPasswordVerify)
	app.Post("/forgot-password", handlers.SimpleAuthMiddleware, handlers.HandleForgotPassword)
	app.Post("/forgot-password/update", handlers.SimpleAuthMiddleware, handlers.HandleForgotPasswordUpdate)

	app.Get("/domains", handlers.AuthMiddleware, handlers.HandleDomainsPage)
	app.Post("/domains/add/new", handlers.AuthMiddleware, handlers.AddNewDomains)
	app.Get("/domains/add", handlers.AuthMiddleware, handlers.AddNewDomainsPage)
	app.Delete("/domains/stop", handlers.AuthMiddleware, handlers.HandleStopMonitoring)
	app.Get("/domains/check", handlers.AuthMiddleware, handlers.HandleCheckDomains)
	app.Get("/domains/more/:id", handlers.AuthMiddleware, handlers.HandleDomainInfoShowPage)

	return app
}
