package user

import (
	"context"

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
