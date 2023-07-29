package api

import (
	"errors"
	"time"

	h "github.com/SaidovZohid/certalert.info/api/handlers"
	"github.com/SaidovZohid/certalert.info/config"
	"github.com/SaidovZohid/certalert.info/pkg/logger"
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
		DisableStartupMessage:   true,
		Views:                   engine,
		EnableTrustedProxyCheck: true,
		WriteTimeout:            10 * time.Minute,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			opt.Log.Error(err)
			return c.Render("errors/500", fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Redirect invalid API requests to the main URL
	app.Use(recover.New(recover.Config{EnableStackTrace: true}))
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		return c.Next()
	})

	app.Use(favicon.New(favicon.Config{
		File: "./static/favicon.ico",
		URL:  "/favicon.ico",
	}))
	app.Static("/static", "./static")

	handlers := h.New(&h.HandlerV1Options{
		Cfg:      opt.Cfg,
		Log:      opt.Log,
		Strg:     opt.Strg,
		InMemory: opt.InMemory,
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
	app.Get("/signup/verify", handlers.SimpleAuthMiddleware, handlers.HandleVerifyUserSignUp)

	// login
	app.Post("/login/user", handlers.SimpleAuthMiddleware, handlers.HandeLoginUser)
	app.Get("/login", handlers.SimpleAuthMiddleware, handlers.HandleGetLoginPage)

	// google login
	app.Get("/login/google", handlers.SimpleAuthMiddleware, handlers.HandleGoogleAuth)
	app.Get("/login/google/callback", handlers.SimpleAuthMiddleware, handlers.HandleGoogleCallback)

	app.Get("/logout", handlers.HandleLogout)

	return app
}
