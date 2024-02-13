package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func (h *handlerV1) HandleGetLandingPage(c *fiber.Ctx) error {
	user, _ := h.getAuth(c)
	if user == nil {
		return c.Render("home/index", nil)
	}

	// user2, err := h.strg.User().GetUserByEmail(context.Background(), user.Email)
	// if err != nil && !errors.Is(err, pgx.ErrNoRows) {
	// 	h.log.Error(err)
	// 	return errors.New("something went unexpected")
	// }

	return c.Render("home/index", fiber.Map{
		"user": user,
		// "id":           user2.ID,
		// "telegram_bot": h.cfg.TelegramBotUsername,
	})
}
