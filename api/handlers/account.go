package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

func (h *handlerV1) HandleAccountPage(c *fiber.Ctx) error {
	payload, _ := h.getAuth(c)

	bind := fiber.Map{}
	bind["user"] = payload

	user, err := h.strg.User().GetUserByEmail(context.Background(), payload.Email)
	if err != nil {
		return err
	}

	bind["maxDomainTracking"] = 5
	if user.MaxDomainsTracking != nil {
		bind["maxDomainTracking"] = user.MaxDomainsTracking
	}

	return c.Render("account/index", bind)
}

func (h *handlerV1) HandleDeleteAccount(c *fiber.Ctx) error {
	payload, _ := h.getAuth(c)

	err := h.strg.User().DeleteUser(context.Background(), payload.UserID)
	if err != nil {
		return err
	}

	return c.Redirect("/signup")
}
