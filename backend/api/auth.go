// Purpose: Login endpoint returns JWT. Uses bcrypt password hash from DB.

package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"radiokpowka/backend/auth"
	"radiokpowka/backend/db"
)

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResp struct {
	Token string `json:"token"`
}

func LoginHandler(deps RouterDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req loginReq
		if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		var u db.User
		if err := deps.DB.Where("username = ?", req.Username).First(&u).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}

		if !db.CheckPassword(u.PasswordHash, req.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		token, err := auth.SignJWT(deps.Cfg.JWTSecret, auth.Claims{
			UserID: u.ID,
			Role:   u.Role,
			Exp:    time.Now().Add(24 * time.Hour).Unix(),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token sign failed"})
			return
		}

		c.JSON(http.StatusOK, loginResp{Token: token})
	}
}

func MeHandler(deps RouterDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := auth.MustGetClaims(c)
		c.JSON(http.StatusOK, gin.H{
			"user_id": claims.UserID,
			"role":    claims.Role,
		})
	}
}
