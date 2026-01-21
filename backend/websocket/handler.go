// Purpose: WS handler with ping/pong, minimal auth.

package websocket

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	userID uuid.UUID
	role   string
	hub    *Hub
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// CORS is handled at HTTP layer; allow for dev.
		return true
	},
}

func ServeWS(w http.ResponseWriter, r *http.Request, jwtSecret string, tokenQuery string, authHeader string, hub *Hub) {
	role := "listener"
	userID := uuid.Nil

	tokenStr := tokenQuery
	if tokenStr == "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
	}
	if tokenStr != "" {
		if uid, rr, ok := parseJWT(jwtSecret, tokenStr); ok {
			userID = uid
			role = rr
		}
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	c := &Client{
		conn:   conn,
		send:   make(chan []byte, 64),
		userID: userID,
		role:   role,
		hub:    hub,
	}

	hub.register <- c

	conn.SetReadLimit(1 << 20)
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	go c.writeLoop()
	c.readLoop()
}

func (c *Client) readLoop() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		// ignore client messages in scaffold
	}
}

func (c *Client) writeLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			_ = c.conn.WriteMessage(websocket.TextMessage, msg)
		case <-ticker.C:
			_ = c.conn.WriteMessage(websocket.PingMessage, []byte("ping"))
		}
	}
}

type jwtClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func parseJWT(secret, tokenStr string) (uuid.UUID, string, bool) {
	tok, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !tok.Valid {
		return uuid.Nil, "", false
	}
	claims, ok := tok.Claims.(*jwtClaims)
	if !ok {
		return uuid.Nil, "", false
	}
	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, "", false
	}
	role := claims.Role
	if role == "" {
		role = "listener"
	}
	return uid, role, true
}
