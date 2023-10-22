package handlers

import (
	"context"
	"fmt"

	"github.com/SaidovZohid/certalert.info/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/sujit-baniya/flash"
)

// AuthMiddleware uses h.getAuth and if payload is nil redirect to login page if not just skip to another
func (h *handlerV1) AuthMiddleware(c *fiber.Ctx) error {
	payload, _ := h.getAuth(c)
	if payload == nil {
		if c.Request().URI().String() != h.cfg.BaseUrl+"/domains/check" {
			return c.Redirect(c.BaseURL() + "/login")
		} else {
			return c.Send([]byte(fmt.Sprintf(htmlCode, "It appears that your session has expired or doesn't exist. You might not be logged in. Please sign in now and refresh the page!")))
		}
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

// getAuth gets token from local storage and extracts user info from access token
func (h *handlerV1) getAuth(c *fiber.Ctx) (*utils.Payload, string) {
	val := c.Cookies(h.cfg.AuthCookieNameCertAlert, "")
	if val == "" {
		return nil, ""
	}

	res, err := h.strg.Session().GetSessionInfoByID(context.Background(), val)
	if err != nil {
		h.log.Error(err)
		c.ClearCookie()
		return nil, ""
	}

	// verify token and extract user info
	payload, err := utils.VerifyToken(h.cfg, res.AccessToken)
	if err != nil {
		h.log.Error(err)
		c.ClearCookie()
		return nil, ""
	}

	return payload, val
}

func WithFlash(c *fiber.Ctx) error {
	values := flash.Get(c)
	c.Locals("flash", values)
	return c.Next()
}
