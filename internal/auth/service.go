package auth

import (
	"context"
	"errors"
	"time"

	"formify/server/internal/user"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	userRepo    user.Repository
	userService *user.Service
	jwtSecret   string
}

func NewService(userRepo user.Repository, userService *user.Service, jwtSecret string) *Service {
	return &Service{
		userRepo:    userRepo,
		userService: userService,
		jwtSecret:   jwtSecret,
	}
}

func (s *Service) Login(ctx context.Context, email, password string) (*user.User, string, error) {
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", errors.New("invalid email or password")
	}

	if u.IsOAuth {
		return nil, "", errors.New("this account uses OAuth login")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, "", errors.New("invalid email or password")
	}

	token, err := s.GenerateJWT(u)
	if err != nil {
		return nil, "", err
	}

	return u, token, nil
}

func (s *Service) GenerateJWT(u *user.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": u.ID,
		"email":   u.Email,
		"name":    u.Name,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
