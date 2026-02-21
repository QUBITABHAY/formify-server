package user

import (
	"context"
	"time"
)

type User struct {
	ID                 int32      `json:"id"`
	Name               string     `json:"name"`
	Email              string     `json:"email"`
	Password           string     `json:"-"`
	OAuthProvider      *string    `json:"oauth_provider,omitempty"`
	OAuthID            *string    `json:"oauth_id,omitempty"`
	IsOAuth            bool       `json:"is_oauth"`
	GoogleAccessToken  *string    `json:"-"`
	GoogleRefreshToken *string    `json:"-"`
	GoogleTokenExpiry  *time.Time `json:"-"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type Repository interface {
	Create(ctx context.Context, user *User) error
	CreateOAuth(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id int32) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByOAuthID(ctx context.Context, provider, oauthID string) (*User, error)
	Update(ctx context.Context, user *User) error
	UpdatePassword(ctx context.Context, id int32, password string) error
	UpdateOAuthTokens(ctx context.Context, userID int32, accessToken, refreshToken string, expiry time.Time) error
}
