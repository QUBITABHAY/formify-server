package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"formify/server/internal/user"
)

const tokenExpiry = 24 * time.Hour

type Service struct {
	jwtSecret string
}

func NewService(jwtSecret string) *Service {
	return &Service{
		jwtSecret: jwtSecret,
	}
}

func (s *Service) GenerateJWT(u *user.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": u.ID,
		"email":   u.Email,
		"name":    u.Name,
		"exp":     time.Now().Add(tokenExpiry).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
