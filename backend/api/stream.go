// Purpose: Audio stream endpoint. Pipes MP3 audio to clients.

package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func StreamHandler(deps RouterDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("stream: start client=%s ua=%s", c.ClientIP(), c.Request.UserAgent())

		flusher, ok := c.Writer.(http.Flusher)
		if !ok {
			c.String(http.StatusInternalServerError, "stream: flusher not supported")
			return
		}
		flusher.Flush()

		// ВАЖНО: сначала решаем “можем ли стримить”, чтобы не отдавать 200 и потом молча падать.
		// Если в твоём Player есть State() — используй её.
		st := deps.Player.State()
		if st.IsPaused {
			c.String(http.StatusConflict, "server paused")
			return
		}
		if st.Current == nil || st.Current.URL == "" {
			c.String(http.StatusNotFound, "no current track")
			return
		}

		// Заголовки + статус + первый flush
		c.Header("Content-Type", "audio/mpeg")
		c.Header("Cache-Control", "no-store")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Status(http.StatusOK)
		flusher.Flush()

		ctx := c.Request.Context()
		if err := deps.Player.StreamTo(ctx, c.Writer, flusher); err != nil {
			// Теперь ты УВИДИШЬ причину, а не “тишину”
			log.Printf("stream: error: %v", err)
			return
		}

		log.Printf("stream: done client=%s", c.ClientIP())
	}
}
