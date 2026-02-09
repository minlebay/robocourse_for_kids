package users

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("invalid token")

type claims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string   `json:"role"`
	jwt.RegisteredClaims
}

func (h *Handler) generateToken(userID uuid.UUID, role string) (string, error) {
	claims := claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 7)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtKey)
}

func (h *Handler) ParseToken(tokenString string) (userID uuid.UUID, role string, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (interface{}, error) {
		return h.jwtKey, nil
	})
	if err != nil {
		return uuid.Nil, "", ErrInvalidToken
	}
	c, ok := token.Claims.(*claims)
	if !ok || !token.Valid {
		return uuid.Nil, "", ErrInvalidToken
	}
	return c.UserID, c.Role, nil
}
