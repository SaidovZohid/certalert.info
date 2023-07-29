package handlers

import "github.com/gofiber/fiber/v2"

func (h *handlerV1) HandleGetLandingPage(c *fiber.Ctx) error {
	user, _ := h.getAuth(c)
	if user == nil {
		return c.Render("home/index", nil)
	}

	return c.Render("home/index", fiber.Map{
		"user": user,
	})
}
