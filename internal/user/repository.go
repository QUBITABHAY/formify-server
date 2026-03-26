package user

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"formify/server/internal/db"
	"formify/server/internal/shared"
)

var ErrUserNotFound = errors.New("user not found")

type repository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) Repository {
	return &repository{queries: queries}
}

func (r *repository) Create(ctx context.Context, user *User) error {
	dbUser, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Name:     user.Name,
		Email:    user.Email,
		Password: user.Password,
	})
	if err != nil {
		return err
	}
	r.mapDBUserToModel(&dbUser, user)
	return nil
}

func (r *repository) CreateOAuth(ctx context.Context, user *User) error {
	provider := ""
	oauthID := ""
	if user.OAuthProvider != nil {
		provider = *user.OAuthProvider
	}
	if user.OAuthID != nil {
		oauthID = *user.OAuthID
	}

	dbUser, err := r.queries.CreateOAuthUser(ctx, db.CreateOAuthUserParams{
		Name:               user.Name,
		Email:              user.Email,
		OauthProvider:      shared.StringToPgtypeText(&provider),
		OauthID:            shared.StringToPgtypeText(&oauthID),
		GoogleAccessToken:  shared.StringToPgtypeText(user.GoogleAccessToken),
		GoogleRefreshToken: shared.StringToPgtypeText(user.GoogleRefreshToken),
		GoogleTokenExpiry:  shared.TimeToPgtypeTimestamptz(user.GoogleTokenExpiry),
	})
	if err != nil {
		return err
	}
	r.mapDBUserToModel(&dbUser, user)
	return nil
}

func (r *repository) GetByID(ctx context.Context, id int32) (*User, error) {
	dbUser, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	user := &User{}
	r.mapDBUserToModel(&dbUser, user)
	return user, nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	dbUser, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	user := &User{}
	r.mapDBUserToModel(&dbUser, user)
	return user, nil
}

func (r *repository) GetByOAuthID(ctx context.Context, provider, oauthID string) (*User, error) {
	dbUser, err := r.queries.GetUserByOAuthID(ctx, db.GetUserByOAuthIDParams{
		OauthProvider: shared.StringToPgtypeText(&provider),
		OauthID:       shared.StringToPgtypeText(&oauthID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	user := &User{}
	r.mapDBUserToModel(&dbUser, user)
	return user, nil
}

func (r *repository) Update(ctx context.Context, user *User) error {
	dbUser, err := r.queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	})
	if err != nil {
		return err
	}
	r.mapDBUserToModel(&dbUser, user)
	return nil
}

func (r *repository) UpdatePassword(ctx context.Context, id int32, password string) error {
	return r.queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:       id,
		Password: password,
	})
}

func (r *repository) UpdateOAuthTokens(ctx context.Context, userID int32, accessToken, refreshToken string, expiry time.Time) error {
	_, err := r.queries.UpdateUserOAuthTokens(ctx, db.UpdateUserOAuthTokensParams{
		ID:                 userID,
		GoogleAccessToken:  shared.StringToPgtypeText(&accessToken),
		GoogleRefreshToken: shared.StringToPgtypeText(&refreshToken),
		GoogleTokenExpiry:  shared.TimeToPgtypeTimestamptz(&expiry),
	})
	return err
}

func (*repository) mapDBUserToModel(dbUser *db.User, user *User) {
	user.ID = dbUser.ID
	user.Name = dbUser.Name
	user.Email = dbUser.Email
	user.Password = dbUser.Password
	user.OAuthProvider = shared.PgtypeTextToString(dbUser.OauthProvider)
	user.OAuthID = shared.PgtypeTextToString(dbUser.OauthID)
	user.IsOAuth = shared.PgtypeBoolToBool(dbUser.IsOauth)
	user.GoogleAccessToken = shared.PgtypeTextToString(dbUser.GoogleAccessToken)
	user.GoogleRefreshToken = shared.PgtypeTextToString(dbUser.GoogleRefreshToken)
	user.GoogleTokenExpiry = shared.PgtypeTimestamptzToTimePtr(dbUser.GoogleTokenExpiry)
	user.CreatedAt = shared.PgtypeTimestamptzToTime(dbUser.CreatedAt)
	user.UpdatedAt = shared.PgtypeTimestamptzToTime(dbUser.UpdatedAt)
}
