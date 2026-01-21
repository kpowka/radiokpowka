// Purpose: JWT signing + claims model.

package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	Exp    int64     `json:"exp"`
}

type jwtClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func SignJWT(secret string, c Claims) (string, error) {
	if secret == "" {
		return "", errors.New("empty secret")
	}
	jc := jwtClaims{
		UserID: c.UserID.String(),
		Role:   c.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(c.Exp, 0)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jc)
	return t.SignedString([]byte(secret))
}
