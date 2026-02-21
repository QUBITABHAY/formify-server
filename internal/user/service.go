package user

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	userRepo Repository
}

func NewService(userRepo Repository) *Service {
	return &Service{userRepo: userRepo}
}

func (s *Service) CreateUser(ctx context.Context, user *User) error {
	if user.Password != "" && !user.IsOAuth {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashedPassword)
	}
	return s.userRepo.Create(ctx, user)
}

func (s *Service) CreateOAuthUser(ctx context.Context, user *User) error {
	return s.userRepo.CreateOAuth(ctx, user)
}

func (s *Service) GetUserByID(ctx context.Context, id int32) (*User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *Service) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

func (s *Service) GetUserByOAuthID(ctx context.Context, provider, oauthID string) (*User, error) {
	return s.userRepo.GetByOAuthID(ctx, provider, oauthID)
}

func (s *Service) UpdateUser(ctx context.Context, user *User) error {
	return s.userRepo.Update(ctx, user)
}

func (s *Service) UpdatePassword(ctx context.Context, id int32, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.userRepo.UpdatePassword(ctx, id, string(hashedPassword))
}

func (s *Service) UpdateOAuthTokens(ctx context.Context, userID int32, accessToken, refreshToken string, expiry time.Time) error {
	return s.userRepo.UpdateOAuthTokens(ctx, userID, accessToken, refreshToken, expiry)
}

func (s *Service) GetUserTokens(ctx context.Context, userID int32) (accessToken, refreshToken string, expiry time.Time, err error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", "", time.Time{}, err
	}

	if user.GoogleAccessToken == nil || *user.GoogleAccessToken == "" {
		return "", "", time.Time{}, nil
	}

	accessToken = *user.GoogleAccessToken
	if user.GoogleRefreshToken != nil {
		refreshToken = *user.GoogleRefreshToken
	}
	if user.GoogleTokenExpiry != nil {
		expiry = *user.GoogleTokenExpiry
	}

	return accessToken, refreshToken, expiry, nil
}
