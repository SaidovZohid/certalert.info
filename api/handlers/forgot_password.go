package handlers

import (
	"context"
	"errors"
	"time"

	apiModels "github.com/SaidovZohid/certalert.info/api/models"
	"github.com/SaidovZohid/certalert.info/pkg/email"
	"github.com/SaidovZohid/certalert.info/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4"
)

func (h *handlerV1) HandleForgotPasswordPage(c *fiber.Ctx) error {
	return c.Render("forgot_password/reset_pw_link", nil)
}

func (h *handlerV1) HandleForgotPassword(c *fiber.Ctx) error {
	var req apiModels.ForgotPasswordReq
	if err := c.BodyParser(&req); err != nil {
		return err
	}

	// email -> link -> link-verify -> open-password-create-web-page -> get-an-new-password -> update user password and redirect-to-login-page

	user, err := h.strg.User().GetUserByEmail(context.Background(), req.Email)
	if err != nil && errors.Is(err, pgx.ErrNoRows) || user == nil {
		h.log.Error(err)
		return c.Render("forgot_password/reset_pw_link", fiber.Map{
			"error": "The provided email address does not exist in our records.",
		})
	}

	_, isOK := h.forgotPasswordUserReq[req.Email]
	if isOK {
		return c.Render("forgot_password/reset_pw_link", fiber.Map{
			"error": "Remember, we've already sent you the password reset link. It remains valid for 15 minutes.",
		})
	}

	token, _, err := utils.CreateToken(h.cfg, &utils.TokenParams{
		UserID:   user.ID,
		Email:    req.Email,
		Duration: h.cfg.ForgotPasswordLinkTokenTime,
	})
	if err != nil {
		h.log.Error(err)
		return c.Render("forgot_password/reset_pw_link", fiber.Map{
			"error": "Try again, failed to create link to update password.",
		})
	}

	ok := true
	var tokenKey string
	for ok {
		tokenKey = utils.GenerateRandomString(12)
		_, ok = h.tokens[tokenKey]
	}
	h.forgotPasswordUserReq[req.Email] = tokenKey
	h.tokens[tokenKey] = TokenDataValidAndToken{
		Token:     token,
		Email:     req.Email,
		ExpiresAt: time.Now().Add(h.cfg.ForgotPasswordLinkTokenTime),
	}

	fullname := user.FirstName + " " + user.LastName
	go func() {
		if err := email.SendEmail(h.cfg, &email.SendEmailRequest{
			To:   []string{req.Email},
			Type: email.ForgotPasswordEmail,
			Body: map[string]string{
				"link": h.cfg.BaseUrl + "/forgot-password/options?token=" + tokenKey,
				"name": fullname,
			},
			Subject: "Password Reset Request",
		}); err != nil {
			h.log.Error(err)
		}
	}()

	return c.Render("forgot_password/reset_pw_link", fiber.Map{
		"success": "Reset password link sent to email address! It remains valid for 15 minutes. Check your inbox and spam.",
	})
}

func (h *handlerV1) HandleForgotPasswordVerify(c *fiber.Ctx) error {
	tokenKey := c.Query("token", "")
	if tokenKey == "" {
		return c.Redirect("/")
	}

	data, ok := h.tokens[tokenKey]
	if !ok {
		return c.Redirect("/")
	}

	if time.Now().After(data.ExpiresAt) {
		delete(h.tokens, tokenKey)
		delete(h.forgotPasswordUserReq, data.Email)
		return errors.New("token is invalid")
	}

	payload, err := utils.VerifyToken(h.cfg, data.Token)
	if err != nil {
		return errors.New("token is invalid")
	}

	user, err := h.strg.User().GetUserByEmail(context.Background(), payload.Email)
	if err != nil && errors.Is(err, pgx.ErrNoRows) || user == nil {
		h.log.Error(err)
		return errors.New("user does not exist please check your email or create a free account in CertAlert")
	}

	return c.Render("forgot_password/reset_pw", fiber.Map{
		"key_prvd_by_rst_link": tokenKey,
	})
}

func (h *handlerV1) HandleForgotPasswordUpdate(c *fiber.Ctx) error {
	tokenKey := c.Query("token", "")
	if tokenKey == "" {
		return c.Redirect("/")
	}

	var req apiModels.UpdatePasswordReq
	if err := c.BodyParser(&req); err != nil {
		h.log.Error(err)
		return errors.New("request body is not acceptable")
	}

	data, ok := h.tokens[tokenKey]
	if !ok {
		return c.Redirect("/")
	}

	if time.Now().After(data.ExpiresAt) {
		delete(h.tokens, tokenKey)
		delete(h.forgotPasswordUserReq, data.Email)
		return errors.New("token is invalid")
	}

	payload, err := utils.VerifyToken(h.cfg, data.Token)
	if err != nil {
		h.log.Error(err)
		return errors.New("token is invalid")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		h.log.Error(err)
		return errors.New("something went unexpected, try again")
	}

	if err = h.strg.User().UpdateUserPassword(context.Background(), payload.UserID, hashedPassword); err != nil {
		h.log.Error(err)
		return errors.New("something wen unexpected, try again")
	}

	delete(h.tokens, tokenKey)
	return c.Redirect("/login")
}
