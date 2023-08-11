package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	apiModels "github.com/SaidovZohid/certalert.info/api/models"
	"github.com/SaidovZohid/certalert.info/pkg/email"
	"github.com/SaidovZohid/certalert.info/pkg/utils"
	"github.com/SaidovZohid/certalert.info/storage/models"
	"github.com/gofiber/fiber/v2"
)

// handles signup page rendering
func (h *handlerV1) HandleGetSignUpPage(c *fiber.Ctx) error {
	return c.Render("signup/index", nil)
}

// handles user sign up page form request
func (h *handlerV1) HandeSignupUser(c *fiber.Ctx) error {
	var req apiModels.SignUpReq
	if err := c.BodyParser(&req); err != nil {
		h.log.Error(err)
		return c.Render("signup/index", fiber.Map{
			"error": "Fill all required fields!",
		})
	}

	splitedName := strings.Split(req.Fullname, " ")
	if len(splitedName) > 2 || len(splitedName) < 2 {
		return c.Render("signup/index", fiber.Map{
			"error": "Enter your fullname!",
		})
	}

	user, err := h.strg.User().GetUserByEmail(context.Background(), req.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) || user != nil {
		h.log.Error(err)
		return c.Render("signup/index", fiber.Map{
			"error": "Email already used.",
		})
	}

	password, err := utils.HashPassword(req.Password)
	if err != nil {
		h.log.Error(err)
		return c.Render("signup/index", fiber.Map{
			"error": "Password is not valid.",
		})
	}

	token, _, err := utils.CreateToken(h.cfg, &utils.TokenParams{
		Email:    req.Email,
		Duration: h.cfg.SignUPLinkTokenTime,
	})
	if err != nil {
		h.log.Error(err)
		return c.Render("signup/index", fiber.Map{
			"error": "Password is not valid.",
		})
	}

	go func() {
		if err := email.SendEmail(h.cfg, &email.SendEmailRequest{
			To:   []string{req.Email},
			Type: email.VerificationEmail,
			Body: map[string]string{
				"link": h.cfg.BaseUrl + "/signup/options?token=" + token,
				"name": req.Fullname,
			},
			Subject: "Verification code to " + req.Fullname,
		}); err != nil {
			h.log.Error(err)
		}
	}()

	redisUser := &apiModels.UserRedis{
		FirstName: splitedName[0],
		LastName:  splitedName[1],
		Email:     req.Email,
		Password:  password,
	}
	userData, err := json.Marshal(redisUser)
	if err != nil {
		h.log.Error(err)
		return c.Render("signup/index", fiber.Map{
			"error": "Try again, something went wrong.",
		})
	}

	err = h.inMemory.Set("user_"+req.Email, string(userData), h.cfg.SignUPLinkTokenTime)
	if err != nil {
		h.log.Error(err)
		return c.Render("signup/index", fiber.Map{
			"error": "Try again, something went wrong.",
		})
	}

	return c.SendString("Verification link sent to " + req.Email)
}

func (h *handlerV1) HandleVerifyUserSignUp(c *fiber.Ctx) error {
	token := c.Query("token", "")
	if token == "" {
		return c.Redirect(h.cfg.BaseUrl)
	}

	payload, err := utils.VerifyToken(h.cfg, token)
	if err != nil {
		return errors.New("token is invalid or already expired")
	}

	user, err := h.strg.User().GetUserByEmail(context.Background(), payload.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) || user != nil {
		h.log.Error(err)
		return errors.New("token invalid or email already verified")
	}

	val, err := h.inMemory.Get("user_" + payload.Email)
	if err != nil {
		return errors.New("token is invalid or already expired")
	}
	var userNotVerified apiModels.UserRedis
	if err := json.Unmarshal([]byte(val), &userNotVerified); err != nil {
		h.log.Error(err)
		return errors.New("try again, something went wrong")
	}

	_, err = h.strg.User().CreateUser(context.Background(), &models.User{
		FirstName: userNotVerified.FirstName,
		LastName:  userNotVerified.LastName,
		Email:     userNotVerified.Email,
		Password:  userNotVerified.Password,
	})
	if err != nil {
		h.log.Error(err)
		return errors.New("try again, something went wrong")
	}

	return c.Redirect("/login")
}

// handles login page rendering
func (h *handlerV1) HandleGetLoginPage(c *fiber.Ctx) error {
	return c.Render("login/index", nil)
}

// handles user login page form request
func (h *handlerV1) HandeLoginUser(c *fiber.Ctx) error {
	var req apiModels.LoginReq
	if err := c.BodyParser(&req); err != nil {
		return c.Render("login/index", fiber.Map{
			"error": "Fill all required fields",
		})
	}

	user, err := h.strg.User().GetUserByEmail(context.Background(), req.Email)
	if err != nil {
		h.log.Error(err)
		return c.Render("login/index", fiber.Map{
			"error": "Email not found",
		})
	}

	if err := utils.CheckPassword(req.Password, user.Password); err != nil {
		return c.Render("login/index", fiber.Map{
			"error": "Email or Passoword is incorrect",
		})
	}

	if err := handleLoginDependencies(c, h, user.ID, &User{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}); err != nil {
		return c.Render("login/index", fiber.Map{
			"error": "Try again, something went wrong.",
		})
	}

	return c.Redirect("/domains", 302)
}

// Google sign in or up redirect user to google
func (h *handlerV1) HandleGoogleAuth(c *fiber.Ctx) error {
	url := h.cfg.Google.Conf.AuthCodeURL("randomstate")

	return c.Redirect(url, 307)
}

// Callback for google get code in query by google and authicated access token and get user data and  handle sign in or sign up
func (h *handlerV1) HandleGoogleCallback(c *fiber.Ctx) error {
	if c.Query("state") != "randomstate" {
		return errors.New("user denied login or signup")
	}

	code := c.Query("code")

	token, err := h.cfg.Google.Conf.Exchange(context.Background(), code)
	if err != nil {
		return err
	}

	data, err := h.getUserInfoFromGoogle(token.AccessToken)
	if err != nil {
		return err
	}

	var id int64
	user, err := h.strg.User().GetUserByEmail(context.Background(), data.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		h.log.Error(err)
		return errors.New("something went unexpected")
	}
	if user == nil {
		// ps := utils.GenerateRandomPassword(16)
		// log.Println(ps)
		ps := "qwert"
		randomPassword, err := utils.HashPassword(ps)
		if err != nil {
			h.log.Error(err)
			return errors.New("something went unexpected")
		}
		userCreated, err := h.strg.User().CreateUser(context.Background(), &models.User{
			FirstName: data.FirstName,
			LastName:  data.LastName,
			Email:     data.Email,
			Password:  randomPassword,
		})
		if err != nil {
			return errors.New("failed to create new user")
		}
		id = userCreated.ID
	} else {
		id = user.ID
	}

	if err := handleLoginDependencies(c, h, id, data); err != nil {
		return err
	}

	return c.Redirect("/domains", 302)
}

// Handle log out from website
func (h *handlerV1) HandleLogout(c *fiber.Ctx) error {
	_, id := h.getAuth(c)
	if id == "" {
		return c.Redirect(c.BaseURL() + "/")
	}
	err := h.strg.Session().DeleteSessionByID(context.Background(), id)
	if err != nil {
		h.log.Error(err)
	}

	// It will set the time on 2009 year that is why cookie will automatically deleted
	h.SetCookie(c, h.cfg.AuthCookieNameCertAlert, "", time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC))

	return c.Redirect(c.BaseURL()+"/login", 302)
}
