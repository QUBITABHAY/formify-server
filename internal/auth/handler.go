package auth

import (
	"net/http"
	"time"

	"formify/server/internal/shared"
	"formify/server/internal/user"

	"github.com/labstack/echo/v5"
)

const CookieName = "token"

type Handler struct {
	service      *Service
	userService  *user.Service
	frontendURL  string
	cookieSecure bool
}

func NewHandler(service *Service, userService *user.Service, frontendURL string, cookieSecure bool) *Handler {
	return &Handler{
		service:      service,
		userService:  userService,
		frontendURL:  frontendURL,
		cookieSecure: cookieSecure,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	User UserData `json:"user"`
}

type UserData struct {
	ID    int32  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (h *Handler) getCookieDomain() string {
	frontendURL := h.frontendURL
	for _, prefix := range []string{"https://", "http://"} {
		if len(frontendURL) > len(prefix) && frontendURL[:len(prefix)] == prefix {
			frontendURL = frontendURL[len(prefix):]
			break
		}
	}

	for i, ch := range frontendURL {
		if ch == ':' || ch == '/' {
			frontendURL = frontendURL[:i]
			break
		}
	}

	if frontendURL == "localhost" || frontendURL == "127.0.0.1" {
		return ""
	}
	return "." + frontendURL
}

func (h *Handler) getSameSite() http.SameSite {
	if h.cookieSecure {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}

func (h *Handler) setTokenCookie(c *echo.Context, token string) {
	cookie := &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		Domain:   h.getCookieDomain(),
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: h.getSameSite(),
		MaxAge:   int(24 * time.Hour / time.Second),
	}
	c.SetCookie(cookie)
}

func (h *Handler) clearTokenCookie(c *echo.Context) {
	cookie := &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		Domain:   h.getCookieDomain(),
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: h.getSameSite(),
		MaxAge:   -1,
	}
	c.SetCookie(cookie)
}

func (h *Handler) Login(c *echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return shared.RespondError(c, http.StatusBadRequest, "Invalid request body")
	}

	if req.Email == "" || req.Password == "" {
		return shared.RespondError(c, http.StatusBadRequest, "Email and password are required")
	}

	user, token, err := h.service.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return shared.RespondError(c, http.StatusUnauthorized, err.Error())
	}

	h.setTokenCookie(c, token)

	return c.JSON(http.StatusOK, AuthResponse{
		User: UserData{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	})
}

func (h *Handler) Logout(c *echo.Context) error {
	h.clearTokenCookie(c)
	return c.JSON(http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

func (h *Handler) Me(c *echo.Context) error {
	userID, ok := c.Get("user_id").(float64)
	if !ok {
		return shared.RespondError(c, http.StatusUnauthorized, "Unauthorized")
	}

	u, err := h.userService.GetUserByID(c.Request().Context(), int32(userID))
	if err != nil {
		return shared.RespondError(c, http.StatusNotFound, "User not found")
	}

	return c.JSON(http.StatusOK, AuthResponse{
		User: UserData{
			ID:    u.ID,
			Name:  u.Name,
			Email: u.Email,
		},
	})
}
