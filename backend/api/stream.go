// Purpose: Audio stream endpoint. No YouTube video on client.
// Strategy:
// - If server paused -> stream silence (so clients don't error-loop).
// - If playing -> spawn ffmpeg against yt-dlp direct audio URL, seek to server position.

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func StreamHandler(deps RouterDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		// audio/mpeg streaming
		c.Header("Content-Type", "audio/mpeg")
		c.Header("Cache-Control", "no-store")
		c.Header("X-Content-Type-Options", "nosniff")

		flusher, ok := c.Writer.(http.Flusher)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}

		ctx := c.Request.Context()
		if err := deps.Player.StreamTo(ctx, c.Writer, flusher); err != nil {
			// If client disconnects, ctx will be canceled: ignore noisy errors
			return
		}
	}
}
