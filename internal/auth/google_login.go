package auth

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"

	"formify/server/internal/shared"
	"formify/server/internal/user"
)

const googleProvider = "google"

func (h *Handler) GoogleLogin(c *echo.Context) error {
	r := c.Request()
	w := c.Response()

	q := r.URL.Query()
	q.Set("provider", googleProvider)
	r.URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(w, r)
	return nil
}

func (h *Handler) GoogleCallback(c *echo.Context) error {
	r := c.Request()
	w := c.Response()

	q := r.URL.Query()
	q.Set("provider", googleProvider)
	r.URL.RawQuery = q.Encode()

	gothUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		return shared.RespondError(c, http.StatusUnauthorized, fmt.Sprintf("OAuth authentication failed: %v", err))
	}

	u, err := h.getUserOrCreateFromOAuth(c, gothUser)
	if err != nil {
		return err
	}

	token, err := h.service.GenerateJWT(u)
	if err != nil {
		return shared.RespondError(c, http.StatusInternalServerError, "Failed to generate token")
	}

	h.setTokenCookie(c, token)

	return c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/auth/callback")
}

func (h *Handler) getUserOrCreateFromOAuth(c *echo.Context, gothUser goth.User) (*user.User, error) {
	ctx := c.Request().Context()
	provider := googleProvider

	// Try to find existing OAuth user
	existingUser, err := h.userService.GetUserByOAuthID(ctx, provider, gothUser.UserID)
	if err != nil && !errors.Is(err, user.ErrUserNotFound) {
		return nil, shared.RespondError(c, http.StatusInternalServerError, "Failed to look up user")
	}

	if existingUser != nil {
		if err := h.updateUserOAuthTokens(c, existingUser.ID, gothUser); err != nil {
			return nil, err
		}
		return existingUser, nil
	}

	existingByEmail, err := h.userService.GetUserByEmail(ctx, gothUser.Email)
	if err != nil && !errors.Is(err, user.ErrUserNotFound) {
		return nil, shared.RespondError(c, http.StatusInternalServerError, "Failed to look up user")
	}

	if existingByEmail != nil {
		return h.linkExistingUserToOAuth(c, existingByEmail, gothUser)
	}

	return h.createNewOAuthUser(c, gothUser)
}

func (h *Handler) updateUserOAuthTokens(c *echo.Context, userID int32, gothUser goth.User) error {
	if gothUser.AccessToken == "" {
		return nil
	}

	ctx := c.Request().Context()
	if err := h.userService.UpdateOAuthTokens(ctx, userID, gothUser.AccessToken, gothUser.RefreshToken, gothUser.ExpiresAt); err != nil {
		return shared.RespondError(c, http.StatusInternalServerError, "Failed to update OAuth tokens")
	}
	return nil
}

func (h *Handler) linkExistingUserToOAuth(c *echo.Context, existingUser *user.User, gothUser goth.User) (*user.User, error) {
	ctx := c.Request().Context()
	provider := googleProvider

	existingUser.OAuthProvider = &provider
	existingUser.OAuthID = &gothUser.UserID
	existingUser.IsOAuth = true

	if err := h.userService.UpdateUser(ctx, existingUser); err != nil {
		return nil, shared.RespondError(c, http.StatusInternalServerError, "Failed to link OAuth account")
	}

	if err := h.updateUserOAuthTokens(c, existingUser.ID, gothUser); err != nil {
		return nil, err
	}

	return existingUser, nil
}

func (h *Handler) createNewOAuthUser(c *echo.Context, gothUser goth.User) (*user.User, error) {
	ctx := c.Request().Context()
	provider := googleProvider

	u := &user.User{
		Name:               gothUser.Name,
		Email:              gothUser.Email,
		OAuthProvider:      &provider,
		OAuthID:            &gothUser.UserID,
		IsOAuth:            true,
		GoogleAccessToken:  &gothUser.AccessToken,
		GoogleRefreshToken: &gothUser.RefreshToken,
		GoogleTokenExpiry:  &gothUser.ExpiresAt,
	}

	if err := h.userService.CreateOAuthUser(ctx, u); err != nil {
		return nil, shared.RespondError(c, http.StatusInternalServerError, "Failed to create user")
	}

	return u, nil
}
