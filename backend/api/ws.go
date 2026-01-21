// Purpose: WebSocket endpoint for realtime events.

package api

import (
	"github.com/gin-gonic/gin"
	"radiokpowka/backend/websocket"
)

func WSHandler(deps RouterDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Auth via query ?token=... or Authorization header (browser WS обычно не ставит header)
		token := c.Query("token")
		authHeader := c.GetHeader("Authorization")
		websocket.ServeWS(c.Writer, c.Request, deps.Cfg.JWTSecret, token, authHeader, deps.Hub)
	}
}
