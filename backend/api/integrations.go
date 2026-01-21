// Purpose: Placeholders for integrations connect endpoints (DonAlerts, DonX).

package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"

	"radiokpowka/backend/db"
)

type connectReq struct {
	Mode string `json:"mode"`
}

func DonAlertsConnectHandler(deps RouterDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req connectReq
		_ = c.ShouldBindJSON(&req)

		now := time.Now().UTC()
		row := db.Integration{
			ID:          uuid.New(),
			Type:        "donationalerts",
			ConfigJSON:  datatypes.JSON([]byte(`{"mode":"placeholder"}`)),
			ConnectedAt: &now,
		}
		_ = deps.DB.Create(&row).Error
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

func DonXConnectHandler(deps RouterDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req connectReq
		_ = c.ShouldBindJSON(&req)

		now := time.Now().UTC()
		row := db.Integration{
			ID:          uuid.New(),
			Type:        "donx",
			ConfigJSON:  datatypes.JSON([]byte(`{"mode":"placeholder"}`)),
			ConnectedAt: &now,
		}
		_ = deps.DB.Create(&row).Error
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}
