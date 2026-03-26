package auth

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/markbates/goth/gothic"

	"formify/server/internal/shared"
	"formify/server/internal/user"
)

func (h *Handler) GoogleLogin(c *echo.Context) error {
	r := c.Request()
	w := c.Response()

	q := r.URL.Query()
	q.Set("provider", "google")
	r.URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(w, r)
	return nil
}

func (h *Handler) GoogleCallback(c *echo.Context) error {
	r := c.Request()
	w := c.Response()

	q := r.URL.Query()
	q.Set("provider", "google")
	r.URL.RawQuery = q.Encode()

	gothUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		return shared.RespondError(c, http.StatusUnauthorized, fmt.Sprintf("OAuth authentication failed: %v", err))
	}

	provider := "google"
	existingUser, err := h.userService.GetUserByOAuthID(c.Request().Context(), provider, gothUser.UserID)
	if err != nil && !errors.Is(err, user.ErrUserNotFound) {
		return shared.RespondError(c, http.StatusInternalServerError, "Failed to look up user")
	}

	var u *user.User

	if existingUser != nil {
		u = existingUser
		if gothUser.AccessToken != "" {
			if err := h.userService.UpdateOAuthTokens(c.Request().Context(), u.ID, gothUser.AccessToken, gothUser.RefreshToken, gothUser.ExpiresAt); err != nil {
				return shared.RespondError(c, http.StatusInternalServerError, "Failed to update OAuth tokens")
			}
		}
	} else {
		existingByEmail, err := h.userService.GetUserByEmail(c.Request().Context(), gothUser.Email)
		if err != nil && !errors.Is(err, user.ErrUserNotFound) {
			return shared.RespondError(c, http.StatusInternalServerError, "Failed to look up user")
		}

		if existingByEmail != nil {
			existingByEmail.OAuthProvider = &provider
			existingByEmail.OAuthID = &gothUser.UserID
			existingByEmail.IsOAuth = true
			if err := h.userService.UpdateUser(c.Request().Context(), existingByEmail); err != nil {
				return shared.RespondError(c, http.StatusInternalServerError, "Failed to link OAuth account")
			}
			if gothUser.AccessToken != "" {
				if err := h.userService.UpdateOAuthTokens(
					c.Request().Context(),
					existingByEmail.ID,
					gothUser.AccessToken,
					gothUser.RefreshToken,
					gothUser.ExpiresAt,
				); err != nil {
					return shared.RespondError(c, http.StatusInternalServerError, "Failed to update OAuth tokens")
				}
			}
			u = existingByEmail
		} else {
			u = &user.User{
				Name:               gothUser.Name,
				Email:              gothUser.Email,
				OAuthProvider:      &provider,
				OAuthID:            &gothUser.UserID,
				IsOAuth:            true,
				GoogleAccessToken:  &gothUser.AccessToken,
				GoogleRefreshToken: &gothUser.RefreshToken,
				GoogleTokenExpiry:  &gothUser.ExpiresAt,
			}
			if err := h.userService.CreateOAuthUser(c.Request().Context(), u); err != nil {
				return shared.RespondError(c, http.StatusInternalServerError, "Failed to create user")
			}
		}
	}

	token, err := h.service.GenerateJWT(u)
	if err != nil {
		return shared.RespondError(c, http.StatusInternalServerError, "Failed to generate token")
	}

	h.setTokenCookie(c, token)

	return c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/auth/callback")
}
