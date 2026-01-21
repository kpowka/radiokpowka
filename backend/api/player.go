// Purpose: Player control and state endpoints.

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetPlayerStateHandler(deps RouterDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, deps.Player.State())
	}
}

func PlayerPlayHandler(deps RouterDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := deps.Player.Play(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func PlayerPauseHandler(deps RouterDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := deps.Player.Pause(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func PlayerNextHandler(deps RouterDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := deps.Player.Next(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func PlayerPrevHandler(deps RouterDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := deps.Player.Prev(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}

type volumeReq struct {
	Volume float64 `json:"volume"`
}

func PlayerVolumeHandler(deps RouterDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req volumeReq
		if err := c.ShouldBindJSON(&req); err != nil || req.Volume < 0 || req.Volume > 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid volume"})
			return
		}
		deps.Player.SetVolume(req.Volume)
		c.Status(http.StatusNoContent)
	}
}
