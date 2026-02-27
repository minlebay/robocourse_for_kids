package users

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("invalid token")

// SlidingRefreshThreshold — если до истечения токена осталось меньше этого времени,
// при следующем запросе в ответ вернётся новый токен (заголовок X-New-Token).
const SlidingRefreshThreshold = 30 * time.Minute

// TokenTTL — время жизни JWT.
const TokenTTL = 1 * time.Hour

type claims struct {
	UserID             uuid.UUID `json:"user_id"`
	Roles              []string  `json:"roles"`
	MustChangePassword bool      `json:"must_change_password"`
	jwt.RegisteredClaims
}

func (h *Handler) generateToken(userID uuid.UUID, roles []string, mustChangePassword bool) (string, error) {
	claims := claims{
		UserID:             userID,
		Roles:              roles,
		MustChangePassword: mustChangePassword,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "learn_kids",
			Subject:   userID.String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtKey)
}

// NewToken выдаёт новый JWT для пользователя (для sliding session).
func (h *Handler) NewToken(userID uuid.UUID, roles []string, mustChangePassword bool) (string, error) {
	return h.generateToken(userID, roles, mustChangePassword)
}

// ParseToken парсит JWT и возвращает userID, roles, mustChangePassword и время истечения токена.
// При невалидном или истёкшем токене возвращает err.
func (h *Handler) ParseToken(tokenString string) (userID uuid.UUID, roles []string, mustChangePassword bool, expiresAt time.Time, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (interface{}, error) {
		return h.jwtKey, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return uuid.Nil, nil, false, time.Time{}, ErrInvalidToken
	}
	c, ok := token.Claims.(*claims)
	if !ok || !token.Valid {
		return uuid.Nil, nil, false, time.Time{}, ErrInvalidToken
	}
	if c.ExpiresAt != nil {
		expiresAt = c.ExpiresAt.Time
	}
	return c.UserID, c.Roles, c.MustChangePassword, expiresAt, nil
}
