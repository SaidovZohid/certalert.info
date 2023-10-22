package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/SaidovZohid/certalert.info/api/models"
	"github.com/SaidovZohid/certalert.info/pkg/email"
	"github.com/SaidovZohid/certalert.info/pkg/utils"
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

func (h *handlerV1) HandleChangeEmailPage(c *fiber.Ctx) error {
	return c.Render("account/change_email", nil)
}

// handles change email request by getting the email and sending link to verify new email addrs
func (h *handlerV1) HandleChangeEmail(c *fiber.Ctx) error {
	payload, _ := h.getAuth(c)

	var req models.ChangeEmailReq
	if err := c.BodyParser(&req); err != nil {
		h.log.Error(err)
		return c.Render("account/change_email", fiber.Map{
			"error": "Please, enter your valid email address!",
		})
	}

	if req.Email == payload.Email {
		return c.Render("account/change_email", fiber.Map{
			"error": "Please, enter your valid email address!",
		})
	}

	val, err := h.inMemory.Get("change_user_email_" + req.Email)
	if err == nil || val != "" {
		h.log.Error(err)
		return c.Render("account/change_email", fiber.Map{
			"error": "We have already sent a link to this email address. The link is valid for 5 minutes.",
		})
	}

	user, err := h.strg.User().GetUserByEmail(context.Background(), req.Email)
	if !errors.Is(err, sql.ErrNoRows) || (user != nil) {
		h.log.Error(err)
		return c.Render("account/change_email", fiber.Map{
			"error": "Email address is already in use",
		})
	}

	token, _, err := utils.CreateToken(h.cfg, &utils.TokenParams{
		UserID:   payload.UserID,
		Email:    req.Email,
		Duration: h.cfg.UpdateEmailLinkTokenTime,
	})
	if err != nil {
		h.log.Error(err)
		return c.Render("account/change_email", fiber.Map{
			"error": "Please, try again after some minute. Our system faced with problem",
		})
	}

	go func() {
		if err := email.SendEmail(h.cfg, &email.SendEmailRequest{
			To:   []string{req.Email},
			Type: email.ChangeEmail,
			Body: map[string]string{
				"link": h.cfg.BaseUrl + "/account/change-email/options?token=" + token,
			},
			Subject: "Change Email Request",
		}); err != nil {
			h.log.Error(err)
		}
	}()

	err = h.inMemory.Set("change_user_email_"+req.Email, req.Email, h.cfg.UpdateEmailLinkTokenTime)
	if err != nil {
		h.log.Error(err)
		return c.Render("account/change_email", fiber.Map{
			"error": "Please, try again after some minute. Our system faced with problem",
		})
	}

	return c.Render("account/change_email", fiber.Map{
		"success": fmt.Sprintf("We have received your request to change your email address to [%v]. We have sent a verification link to this new email address for confirmation.", req.Email),
	})
}

func (h *handlerV1) HandleChangeEmailOptions(c *fiber.Ctx) error {
	token := c.Query("token", "")
	if token == "" {
		return c.Redirect(h.cfg.BaseUrl)
	}

	payload, err := utils.VerifyToken(h.cfg, token)
	if err != nil {
		h.log.Error(err)
		h.log.Info("User input -> " + token)
		return errors.New("token is invalid or already expired")
	}

	val, _ := h.inMemory.Get("change_user_email_" + payload.Email)
	if val == "" {
		return errors.New("token is invalid or already expired")
	}

	err = h.strg.User().UpdateUserEmail(context.Background(), payload.UserID, payload.Email)
	if err != nil {
		h.log.Error(err)
		return errors.New("please, try again after some minute. Our system faced with problem")
	}

	err = h.strg.Session().DeleteSessionByUserID(context.Background(), payload.UserID)
	if err != nil {
		h.log.Error(err)
		return errors.New("please, try again after some minute. Our system faced with problem")
	}

	err = h.inMemory.Del("change_user_email_" + payload.Email)
	if err != nil {
		h.log.Error(err)
		return errors.New("please, try again after some minute. Our system faced with problem")
	}

	return c.Redirect("/login")
}
