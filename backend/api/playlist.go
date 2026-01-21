// Purpose: Playlist/queue list + add track.

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"radiokpowka/backend/auth"
)

type playlistAddReq struct {
	URL        string `json:"url"`
	AddedByNick string `json:"added_by_nick,omitempty"`
	InsertNext bool   `json:"insert_next,omitempty"`
	IsDonation bool   `json:"is_donation,omitempty"`
}

func PlaylistListHandler(deps RouterDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		items, err := deps.Player.ListQueue()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "queue list failed"})
			return
		}
		c.JSON(http.StatusOK, items)
	}
}

func PlaylistAddHandler(deps RouterDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req playlistAddReq
		if err := c.ShouldBindJSON(&req); err != nil || req.URL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		pc := auth.MustGetOptionalClaims(c)

		addedByNick := req.AddedByNick
		addedByUser := pc.UserIDPtr()
		role := pc.Role

		// If authenticated owner -> ignore client nick and use username/role (backend can improve later)
		// For now: allow passing nick for donation/mod flow.
		if role == "owner" && addedByNick == "" {
			addedByNick = "owner"
		}
		if addedByNick == "" {
			addedByNick = "guest"
		}

		_, err := deps.Player.AddTrack(req.URL, addedByUser, addedByNick, req.InsertNext, req.IsDonation)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusNoContent)
	}
}
