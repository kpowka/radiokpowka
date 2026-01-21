// Purpose: Gin middleware to validate JWT and attach claims to context.

package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const ctxKey = "rk_claims"

type ParsedClaims struct {
	UserID uuid.UUID
	Role   string
}

func JWTMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		tokenStr := strings.TrimPrefix(h, "Bearer ")

		tok, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(token *jwt.Token) (any, error) {
			return []byte(secret), nil
		})
		if err != nil || !tok.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := tok.Claims.(*jwtClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			return
		}

		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
			return
		}

		c.Set(ctxKey, ParsedClaims{UserID: uid, Role: claims.Role})
		c.Next()
	}
}

func MustGetClaims(c *gin.Context) ParsedClaims {
	v, ok := c.Get(ctxKey)
	if !ok {
		return ParsedClaims{}
	}
	pc, _ := v.(ParsedClaims)
	return pc
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		pc := MustGetClaims(c)
		if pc.Role != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}
// OptionalJWT parses token if present and sets optional claims, but does not abort when missing/invalid.
// Useful for endpoints that are "public or authenticated".
func OptionalJWT(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		tokenStr := ""

		if strings.HasPrefix(h, "Bearer ") {
			tokenStr = strings.TrimPrefix(h, "Bearer ")
		}

		// also allow ?token=
		if tokenStr == "" {
			if q := c.Query("token"); q != "" {
				tokenStr = q
			}
		}

		if tokenStr == "" {
			c.Set("rk_optional_claims", OptionalClaims{Role: "listener"})
			c.Next()
			return
		}

		tok, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(token *jwt.Token) (any, error) {
			return []byte(secret), nil
		})
		if err != nil || !tok.Valid {
			c.Set("rk_optional_claims", OptionalClaims{Role: "listener"})
			c.Next()
			return
		}

		claims, ok := tok.Claims.(*jwtClaims)
		if !ok {
			c.Set("rk_optional_claims", OptionalClaims{Role: "listener"})
			c.Next()
			return
		}

		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.Set("rk_optional_claims", OptionalClaims{Role: "listener"})
			c.Next()
			return
		}

		c.Set("rk_optional_claims", OptionalClaims{UserID: uid, Role: claims.Role})
		c.Next()
	}
}

type OptionalClaims struct {
	UserID uuid.UUID
	Role   string
}

func MustGetOptionalClaims(c *gin.Context) OptionalClaims {
	v, ok := c.Get("rk_optional_claims")
	if !ok {
		return OptionalClaims{Role: "listener"}
	}
	oc, _ := v.(OptionalClaims)
	if oc.Role == "" {
		oc.Role = "listener"
	}
	return oc
}

func (o OptionalClaims) UserIDPtr() *uuid.UUID {
	if o.UserID == uuid.Nil {
		return nil
	}
	u := o.UserID
	return &u
}
