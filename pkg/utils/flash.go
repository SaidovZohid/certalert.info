package utils

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v2"
)

type FlashMessage struct {
	Type    string `json:"type"`
	Key     string `json:"key"`
	Message string `json:"message"`
}

func SetFlash(c *fiber.Ctx, messageType, key, message string) {
	flash := FlashMessage{
		Type:    messageType,
		Key:     key,
		Message: message,
	}

	flashJSON, err := json.Marshal(flash)
	if err != nil {
		fmt.Println("Could not set flash message:", err)
		return
	}

	c.Cookie(&fiber.Cookie{
		Name:     "flash",
		Value:    url.QueryEscape(string(flashJSON)),
		Path:     "/",
		HTTPOnly: true,
		Expires:  time.Now().Add(5 * time.Second), // Set an appropriate expiration time
	})
}

func GetFlash(c *fiber.Ctx) *FlashMessage {
	flashCookie := c.Cookies("flash")

	flashValue, err := url.QueryUnescape(flashCookie)
	if err != nil {
		return nil
	}

	flash := FlashMessage{}
	if err := json.Unmarshal([]byte(flashValue), &flash); err != nil {
		return nil
	}

	// Clear flash cookie
	c.Cookie(&fiber.Cookie{
		Name:   "flash",
		Path:   "/",
		MaxAge: -1,
	})

	return &flash
}
