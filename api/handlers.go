package api

import "github.com/gofiber/fiber/v2"

func HandleError(c *fiber.Ctx, err error) error {
	return c.Render("errors/500", fiber.Map{
		"error": err.Error(),
	})
}
	