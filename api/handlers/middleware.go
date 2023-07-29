package handlers

import (
	"context"

	"github.com/SaidovZohid/certalert.info/pkg/utils"
	"github.com/gofiber/fiber/v2"
)

// func uses h.getAuth and if payload is nil redirect to login page if not just skip to another
func (h *handlerV1) AuthMiddleware(c *fiber.Ctx) error {
	payload, _ := h.getAuth(c)
	if payload == nil {
		return c.Redirect(c.BaseURL() + "/login")
	}

	return c.Next()
}

func (h *handlerV1) SimpleAuthMiddleware(c *fiber.Ctx) error {
	payload, _ := h.getAuth(c)
	if payload != nil {
		return c.Redirect(h.cfg.BaseUrl+"/", 301)
	}

	return c.Next()
}

// func gets token from local storage and extracts user info from access token
func (h *handlerV1) getAuth(c *fiber.Ctx) (*utils.Payload, string) {
	val := c.Cookies(h.cfg.AuthCookieNameCertAlert, "")
	if val == "" {
		return nil, ""
	}

	res, err := h.strg.Session().GetSessionInfoByID(context.Background(), val)
	if err != nil {
		c.ClearCookie()
		return nil, ""
	}

	// verify token and extract user info
	payload, err := utils.VerifyToken(h.cfg, res.AccessToken)
	if err != nil {
		c.ClearCookie()
		return nil, ""
	}

	return payload, val
}
