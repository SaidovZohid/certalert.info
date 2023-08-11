package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/SaidovZohid/certalert.info/api/models"
	"github.com/gofiber/fiber/v2"
)

func (h *handlerV1) AddNewDomainsPage(c *fiber.Ctx) error {
	payload, _ := h.getAuth(c)

	bind := fiber.Map{}
	bind["user"] = payload

	return c.Render("domains/add", bind)
}

func (h *handlerV1) AddNewDomains(c *fiber.Ctx) error {
	var req models.DomainsReq
	if err := c.BodyParser(&req); err != nil {
		return err
	}

	// Split the input string into individual words
	domains := strings.Fields(req.Domains)
	payload, _ := h.getAuth(c)
	if payload == nil {
		return c.Redirect(c.BaseURL() + "/login")
	}

	err := TrackDomainsAdded(&TrackDomainAdd{
		UserID:  payload.UserID,
		Domains: domains,
		Log:     &h.log,
		Strg:    h.strg,
	})
	if err != nil {
		return err
	}

	return c.Redirect("/domains")
}

func (h *handlerV1) HandleDomainsPage(c *fiber.Ctx) error {
	payload, _ := h.getAuth(c)

	bind := fiber.Map{}
	bind["user"] = payload

	resp, err := h.strg.Domain().GetDomainsWithUserID(context.Background(), payload.UserID)
	if err != nil {
		return err
	}
	if len(resp) > 0 {
		bind["domains"] = resp
		if len(resp) == 1 {
			bind["domainsHas"] = fmt.Sprintf("%v domain", len(resp))
		} else {
			bind["domainsHas"] = fmt.Sprintf("%v domains", len(resp))
		}
	}

	return c.Render("domains/domains", bind)
}

func (h *handlerV1) HandleStopMonitoring(c *fiber.Ctx) error {
	var req models.DomainsNewReq
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	payload, _ := h.getAuth(c)

	if len(req.Domains) == 0 {
		return errors.New("no domain to delete")
	}

	err := h.strg.Domain().DeleteTrackingDomains(context.Background(), payload.UserID, req.Domains)
	if err != nil {
		h.log.Error(err)
		return err
	}

	return c.Send([]byte("deleted"))
}

func (h *handlerV1) HandleCheckDomains(c *fiber.Ctx) error {
	payload, _ := h.getAuth(c)

	user, err := h.strg.User().GetUserByEmail(context.Background(), payload.Email)
	if err != nil {
		h.log.Error(err)
		return err
	}

	if user.LastPollAt != nil {
		time.Sleep(time.Second)
		currentTime := time.Now()

		timeSinceLastPoll := currentTime.Sub(*user.LastPollAt)

		if timeSinceLastPoll < 30*time.Minute {
			return c.Send([]byte(fmt.Sprintf(htmlCode, "You will be able to perform domain checks again after a 30-minute interval from your last check.")))
		}
	}

	domains, err := h.strg.Domain().GetDomainsWithUserID(context.Background(), payload.UserID)
	if err != nil {
		return err
	}

	allDomains := make([]string, 0)
	for _, v := range domains {
		allDomains = append(allDomains, v.DomainName)
	}
	if len(allDomains) == 0 {
		return c.Send([]byte("Currently, there are no domains available for tracking"))
	}

	err = h.CheckExistingDomains(payload.UserID, allDomains)
	if err != nil {
		return err
	}

	return c.Send([]byte(fmt.Sprintf(htmlCode, "Domain checks have been successfully completed. Please refresh the page to view the updated results!")))
}

// TODO: handle rendering domain info page!
func (h *handlerV1) HandleDomainInfoShowPage(c *fiber.Ctx) error {
	id := c.Params("id", "")

	return c.Send([]byte("You are here " + id))
}
