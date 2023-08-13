package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/SaidovZohid/certalert.info/api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/sujit-baniya/flash"
)

func (h *handlerV1) AddNewDomainsPage(c *fiber.Ctx) error {
	payload, _ := h.getAuth(c)

	bind := fiber.Map{}
	bind["user"] = payload

	domains, err := h.strg.Domain().GetDomainsWithUserID(context.Background(), payload.UserID)
	if err != nil {
		h.log.Error(err)
		return flash.WithData(c, fiber.Map{
			"error": "An error occurred. Please try again later or contact support if the issue persists.",
		}).Redirect("/domains")
	}
	user, err := h.strg.User().GetUserByEmail(context.Background(), payload.Email)
	if err != nil {
		h.log.Error(err)
		return flash.WithData(c, fiber.Map{
			"error": "An error occurred. Please try again later or contact support if the issue persists.",
		}).Redirect("/domains")
	}

	if user.MaxDomainsTracking == nil {
		if len(domains) >= 5 {
			return flash.WithData(c, fiber.Map{
				"maxTrackingDomainsExited": "Sorry, you have reached the maximum limit of domains that can be tracked (5 domains). Please remove some domains if you wish to add more, or contact us on Telegram at @zohid_0212 to discuss upgrading your plan. We're here to assist you!",
			}).Redirect("/domains")
		}
	} else {
		if len(domains) >= *user.MaxDomainsTracking {
			return flash.WithData(c, fiber.Map{
				"maxTrackingDomainsExited": fmt.Sprintf("Sorry, you have reached the maximum limit of domains that can be tracked (%v domains). Please remove some domains if you wish to add more, or contact us on Telegram at @zohid_0212 to discuss upgrading your plan. We're here to assist you!", *user.MaxDomainsTracking),
			}).Redirect("/domains")
		}
	}

	return c.Render("domains/add", bind)
}

func (h *handlerV1) AddNewDomains(c *fiber.Ctx) error {
	data := fiber.Map{}
	var req models.DomainsReq
	if err := c.BodyParser(&req); err != nil {
		data["error"] = "Please provide the domains as mentioned above."
		return flash.WithData(c, data).Redirect("/domains/add")
	}

	payload, _ := h.getAuth(c)

	// Split the input string into individual words
	domains := strings.Fields(req.Domains)
	if len(domains) == 0 {
		data["error"] = "Please provide at least one domain to begin tracking. To track domains, use the following format: domain1.com domain2.com domain3.com. Separate each domain with a space."
		return flash.WithData(c, data).Redirect("/domains/add")
	}

	var areAllValidDomain bool
	notValidDomainNames := make([]string, 0)
	for _, v := range domains {
		if !isValidDomain(v) {
			areAllValidDomain = true
			notValidDomainNames = append(notValidDomainNames, v)
		}
	}
	if areAllValidDomain {
		data["error"] = fmt.Sprintf("Please note that the following domain name(s) provided are not valid: %v. Please ensure you enter valid domain names for tracking.", notValidDomainNames)
		data["domains"] = req.Domains
		return flash.WithData(c, data).Redirect("/domains/add")
	}

	user, err := h.strg.User().GetUserByEmail(context.Background(), payload.Email)
	if err != nil {
		h.log.Error(err)
		data["error"] = "An error occurred. Please try again later or contact support if the issue persists."
		return flash.WithData(c, data).Redirect("/domains/add")
	}

	trackingDomains, err := h.strg.Domain().GetDomainsWithUserID(context.Background(), payload.UserID)
	if err != nil {
		h.log.Error(err)
		data["error"] = "An error occurred. Please try again later or contact support if the issue persists."
		return flash.WithData(c, data).Redirect("/domains/add")
	}

	domainsToTrack := len(trackingDomains) + len(domains)
	if user.MaxDomainsTracking == nil {
		if domainsToTrack > 5 {
			data["maxTrackingDomainsExited"] = fmt.Sprintf("Your domain tracking limit is 5, and you currently have %v domains being tracked. You tried to add %v more domains, but the total would exceed the limit. Please remove some domains or contact us on Telegram at @zohid_0212 to discuss upgrading your plan. We're here to assist you!", len(trackingDomains), len(domains))
			data["domains"] = req.Domains
			return flash.WithData(c, data).Redirect("/domains/add")
		}
	} else {
		if domainsToTrack > *user.MaxDomainsTracking {
			data["maxTrackingDomainsExited"] = fmt.Sprintf("Your domain tracking limit is %v, and you currently have %v domains being tracked. You tried to add %v more domains, but the total would exceed the limit. Please remove some domains or contact us on Telegram at @zohid_0212 to discuss upgrading your plan. We're here to assist you!", *user.MaxDomainsTracking, len(trackingDomains), len(domains))
			data["domains"] = req.Domains
			return flash.WithData(c, data).Redirect("/domains/add")
		}
	}

	err = TrackDomainsAdded(&TrackDomainAdd{
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

	trackingDomains, err := h.strg.Domain().GetDomainsWithUserID(context.Background(), payload.UserID)
	if err != nil {
		h.log.Error(err)
		return err
	}

	if len(trackingDomains) == 0 {
		err := h.strg.User().UpdateUserLastPollToNULL(context.Background(), payload.UserID)
		if err != nil {
			return err
		}
	}

	return c.Send([]byte("deleted"))
}

func (h *handlerV1) HandleCheckDomains(c *fiber.Ctx) error {
	payload, _ := h.getAuth(c)

	user, err := h.strg.User().GetUserByEmail(context.Background(), payload.Email)
	if err != nil {
		h.log.Error(err)
		return c.Send([]byte(fmt.Sprintf(htmlCode, "User not found in our records! Please sign up or sign in to your existing account.")))
	}

	if user.LastPollAt != nil {
		time.Sleep(time.Second)
		currentTime := time.Now()

		timeSinceLastPoll := currentTime.Sub(*user.LastPollAt)

		if timeSinceLastPoll < 30*time.Minute {
			return c.Send([]byte(fmt.Sprintf(htmlCode, "You'll be able to perform domain checks again after a 30-minute interval from your last check.")))
		}
	}

	domains, err := h.strg.Domain().GetDomainsWithUserID(context.Background(), payload.UserID)
	if err != nil {
		return c.Send([]byte(fmt.Sprintf(htmlCode, "We encountered an error while retrieving domain information. Please try again.")))
	}

	allDomains := make([]string, 0)
	for _, v := range domains {
		allDomains = append(allDomains, v.DomainName)
	}
	if len(allDomains) == 0 {
		return c.Send([]byte("Currently, there are no domains available for tracking."))
	}

	err = h.CheckExistingDomains(payload.UserID, allDomains)
	if err != nil {
		return c.Send([]byte("We encountered an error while retrieving domain information. Please try again."))
	}

	return c.Send([]byte(fmt.Sprintf(htmlCode, "Domain checks have been successfully completed. Please refresh the page to view the updated results!")))
}

// TODO: handle rendering domain info page!
func (h *handlerV1) HandleDomainInfoShowPage(c *fiber.Ctx) error {
	id := c.Params("id", "")

	payload, _ := h.getAuth(c)

	bind := fiber.Map{}
	bind["user"] = payload
	bind["id"] = id

	return c.Render("domains/info", bind)
}
