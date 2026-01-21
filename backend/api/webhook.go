// Purpose: Donations webhook receiver (DonAlerts/DonX placeholder payload).
// Security: if WEBHOOK_SECRET set -> require header X-RK-Webhook-Secret.

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"radiokpowka/backend/donations"
)

func DonationWebhookHandler(deps RouterDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		if deps.Cfg.WebhookSecret != "" {
			if c.GetHeader("X-RK-Webhook-Secret") != deps.Cfg.WebhookSecret {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid webhook secret"})
				return
			}
		}

		var payload donations.IncomingDonation
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		if err := donations.HandleDonation(c.Request.Context(), donations.Deps{
			DB:     deps.DB,
			Player: deps.Player,
			Hub:    deps.Hub,
		}, payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusNoContent)
	}
}
