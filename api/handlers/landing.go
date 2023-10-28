package handlers

import (
	"context"
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4"
)

func (h *handlerV1) HandleGetLandingPage(c *fiber.Ctx) error {
	user, _ := h.getAuth(c)
	if user == nil {
		return c.Render("home/index", nil)
	}

	user2, err := h.strg.User().GetUserByEmail(context.Background(), user.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		h.log.Error(err)
		return errors.New("something went unexpected")
	}

	log.Println(user2.ID)
	return c.Render("home/index", fiber.Map{
		"user": user,
		"id":   user2.ID,
	})
}
