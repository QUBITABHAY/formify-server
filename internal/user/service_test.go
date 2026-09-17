package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var errDB = errors.New("db error")

type mockRepository struct {
	updateOAuthTokensFunc func(ctx context.Context, userID int32, accessToken, refreshToken string, expiry time.Time) error
	getByIDFunc           func(ctx context.Context, id int32) (*User, error)
	createOAuthFunc       func(ctx context.Context, user *User) error
	updatePasswordFunc    func(ctx context.Context, id int32, password string) error
	getByEmailFunc        func(ctx context.Context, email string) (*User, error)
	createFunc            func(ctx context.Context, user *User) error
	getByOAuthIDFunc      func(ctx context.Context, provider, oauthID string) (*User, error)
	updateFunc            func(ctx context.Context, user *User) error
}

func (m *mockRepository) Create(ctx context.Context, user *User) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, user)
	}
	return nil
}

func (m *mockRepository) CreateOAuth(ctx context.Context, user *User) error {
	if m.createOAuthFunc != nil {
		return m.createOAuthFunc(ctx, user)
	}
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id int32) (*User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	if m.getByEmailFunc != nil {
		return m.getByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *mockRepository) GetByOAuthID(ctx context.Context, provider, oauthID string) (*User, error) {
	if m.getByOAuthIDFunc != nil {
		return m.getByOAuthIDFunc(ctx, provider, oauthID)
	}
	return nil, nil
}

func (m *mockRepository) Update(ctx context.Context, user *User) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, user)
	}
	return nil
}

func (m *mockRepository) UpdatePassword(ctx context.Context, id int32, password string) error {
	if m.updatePasswordFunc != nil {
		return m.updatePasswordFunc(ctx, id, password)
	}
	return nil
}

func (m *mockRepository) UpdateOAuthTokens(ctx context.Context, userID int32, accessToken, refreshToken string, expiry time.Time) error {
	if m.updateOAuthTokensFunc != nil {
		return m.updateOAuthTokensFunc(ctx, userID, accessToken, refreshToken, expiry)
	}
	return nil
}

func TestService_CreateUser(t *testing.T) {
	t.Run("hashes password for regular user", func(t *testing.T) {
		var passedUser *User
		mock := &mockRepository{
			createFunc: func(_ context.Context, u *User) error {
				passedUser = u
				u.ID = 1
				return nil
			},
		}

		svc := NewService(mock)
		u := &User{
			Name:     "Alice",
			Email:    "alice@example.com",
			Password: "plainPassword123",
			IsOAuth:  false,
		}

		err := svc.CreateUser(context.Background(), u)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if passedUser == nil {
			t.Fatal("expected user to be passed to repo")
		}
		if passedUser.Password == "plainPassword123" {
			t.Fatal("password should have been hashed")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(passedUser.Password), []byte("plainPassword123")); err != nil {
			t.Fatalf("hashed password does not match: %v", err)
		}
	})

	t.Run("does not hash password if empty or is OAuth", func(t *testing.T) {
		mock := &mockRepository{
			createFunc: func(_ context.Context, _ *User) error {
				return nil
			},
		}

		svc := NewService(mock)
		u := &User{
			Name:     "OAuth User",
			Email:    "oauth@example.com",
			Password: "",
			IsOAuth:  true,
		}

		err := svc.CreateUser(context.Background(), u)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u.Password != "" {
			t.Fatalf("expected empty password, got %q", u.Password)
		}
	})
}

func TestService_CreateOAuthUser(t *testing.T) {
	var called bool
	mock := &mockRepository{
		createOAuthFunc: func(_ context.Context, _ *User) error {
			called = true
			return nil
		},
	}

	svc := NewService(mock)
	err := svc.CreateOAuthUser(context.Background(), &User{Name: "OAuth User"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected CreateOAuth to be called")
	}
}

func TestService_GetUserByID(t *testing.T) {
	mock := &mockRepository{
		getByIDFunc: func(_ context.Context, id int32) (*User, error) {
			if id == 42 {
				return &User{ID: 42, Name: "Test User"}, nil
			}
			return nil, ErrUserNotFound
		},
	}

	svc := NewService(mock)
	u, err := svc.GetUserByID(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Name != "Test User" {
		t.Fatalf("expected Test User, got %s", u.Name)
	}

	_, err = svc.GetUserByID(context.Background(), 99)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestService_GetUserByEmail(t *testing.T) {
	mock := &mockRepository{
		getByEmailFunc: func(_ context.Context, email string) (*User, error) {
			if email == "test@test.com" {
				return &User{ID: 1, Email: email}, nil
			}
			return nil, ErrUserNotFound
		},
	}

	svc := NewService(mock)
	u, err := svc.GetUserByEmail(context.Background(), "test@test.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Email != "test@test.com" {
		t.Fatalf("expected email match, got %s", u.Email)
	}
}

func TestService_GetUserByOAuthID(t *testing.T) {
	mock := &mockRepository{
		getByOAuthIDFunc: func(_ context.Context, provider, oauthID string) (*User, error) {
			return &User{ID: 5, OAuthProvider: &provider, OAuthID: &oauthID}, nil
		},
	}

	svc := NewService(mock)
	u, err := svc.GetUserByOAuthID(context.Background(), "google", "12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *u.OAuthProvider != "google" || *u.OAuthID != "12345" {
		t.Fatal("mismatch in provider or oauthID")
	}
}

func TestService_UpdateUser(t *testing.T) {
	var called bool
	mock := &mockRepository{
		updateFunc: func(_ context.Context, _ *User) error {
			called = true
			return nil
		},
	}

	svc := NewService(mock)
	err := svc.UpdateUser(context.Background(), &User{ID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected Update to be called")
	}
}

func TestService_UpdatePassword(t *testing.T) {
	var savedHash string
	mock := &mockRepository{
		updatePasswordFunc: func(_ context.Context, _ int32, password string) error {
			savedHash = password
			return nil
		},
	}

	svc := NewService(mock)
	err := svc.UpdatePassword(context.Background(), 1, "newSecretPass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(savedHash), []byte("newSecretPass")); err != nil {
		t.Fatalf("saved password hash does not match newSecretPass: %v", err)
	}
}

func TestService_UpdateOAuthTokens(t *testing.T) {
	var capturedToken string
	mock := &mockRepository{
		updateOAuthTokensFunc: func(_ context.Context, _ int32, accessToken, _ string, _ time.Time) error {
			capturedToken = accessToken
			return nil
		},
	}

	svc := NewService(mock)
	err := svc.UpdateOAuthTokens(context.Background(), 1, "access-123", "refresh-123", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedToken != "access-123" {
		t.Fatalf("expected access-123, got %q", capturedToken)
	}
}

func TestService_GetUserTokens(t *testing.T) {
	now := time.Now()
	token := "token-val"
	refresh := "refresh-val"

	t.Run("returns tokens when present", func(t *testing.T) {
		mock := &mockRepository{
			getByIDFunc: func(_ context.Context, id int32) (*User, error) {
				return &User{
					ID:                 id,
					GoogleAccessToken:  &token,
					GoogleRefreshToken: &refresh,
					GoogleTokenExpiry:  &now,
				}, nil
			},
		}

		svc := NewService(mock)
		acc, ref, exp, err := svc.GetUserTokens(context.Background(), 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if acc != token || ref != refresh || !exp.Equal(now) {
			t.Fatalf("tokens mismatch: got %v, %v, %v", acc, ref, exp)
		}
	})

	t.Run("returns empty when access token is nil or empty", func(t *testing.T) {
		mock := &mockRepository{
			getByIDFunc: func(_ context.Context, id int32) (*User, error) {
				return &User{ID: id}, nil
			},
		}

		svc := NewService(mock)
		acc, ref, exp, err := svc.GetUserTokens(context.Background(), 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if acc != "" || ref != "" || !exp.IsZero() {
			t.Fatalf("expected empty tokens, got acc=%q, ref=%q, exp=%v", acc, ref, exp)
		}
	})

	t.Run("returns error when GetByID fails", func(t *testing.T) {
		mock := &mockRepository{
			getByIDFunc: func(_ context.Context, _ int32) (*User, error) {
				return nil, errDB
			},
		}

		svc := NewService(mock)
		_, _, _, err := svc.GetUserTokens(context.Background(), 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
