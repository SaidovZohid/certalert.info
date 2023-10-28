package utils

import (
	"context"
	"time"

	"github.com/SaidovZohid/certalert.info/config"
	"github.com/SaidovZohid/certalert.info/pkg/logger"
	"github.com/SaidovZohid/certalert.info/pkg/ssl"
	"github.com/SaidovZohid/certalert.info/storage"
)

// UpdateDomainInformationRegularly fetches information about domains at regular intervals
// of x minutes, pausing in between iterations.
func UpdateDomainInformationRegularly(ctx context.Context, strg storage.StorageI, log logger.Logger, cfg *config.Config) {
	// Create a ticker that triggers every 7 minutes
	ticker := time.NewTicker(cfg.PullUpdateDomainInterval)

	// Defer the stopping of the ticker to avoid leaks
	defer ticker.Stop()

	// Run the domain information update loop
	for {
		select {
		case <-ctx.Done():
			log.Info("UpdateDomainInformationRegularly has been canceled!")
			return // If the context is canceled, exit the function
		case <-ticker.C:
			log.Println("Domain information update started")
			// Perform the domain information update logic here
			// Fetch information about various domains
			domains, err := strg.Domain().GetListofDomainsThatExists(ctx)
			if err != nil {
				log.Error(err)
				continue
			}

			for _, domain := range domains {
				domainInfo, err := ssl.PollDomain(ctx, domain)
				if err != nil {
					log.Error("Failed to poll domain", domain, ":", err)
					continue
				}
				err = strg.Domain().UpdateAllTheSameDomainsInfo(ctx, &ssl.DomainTracking{
					DomainName:         domain,
					TrackingDomainInfo: *domainInfo,
				})
				if err != nil {
					log.Error("Failed to update domain", domain, "information in the database:", err)
				}
			}
			log.Info("Successfully pulled info and updated!")
		}
	}
}
