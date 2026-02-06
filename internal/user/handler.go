package user

import (
	"net/http"
	"strconv"

	"formify/server/internal/shared"

	"github.com/labstack/echo/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID    int32  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (h *Handler) CreateUser(c *echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return shared.RespondError(c, http.StatusBadRequest, "Invalid request body")
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		return shared.RespondError(c, http.StatusBadRequest, "Name, email and password are required")
	}

	user := &User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	if err := h.service.CreateUser(c.Request().Context(), user); err != nil {
		return shared.RespondError(c, http.StatusInternalServerError, "Failed to create user")
	}

	return c.JSON(http.StatusCreated, UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	})
}

func (h *Handler) GetUser(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return shared.RespondError(c, http.StatusBadRequest, "Invalid user ID")
	}

	user, err := h.service.GetUserByID(c.Request().Context(), int32(id))
	if err != nil {
		return shared.RespondError(c, http.StatusNotFound, "User not found")
	}

	return c.JSON(http.StatusOK, UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	})
}
