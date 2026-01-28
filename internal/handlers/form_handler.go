package handlers

import (
	"net/http"
	"strconv"
	"time"

	"formify/server/internal/domain"

	"github.com/labstack/echo/v5"
)

type CreateFormRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	UserID      int32   `json:"user_id"`
}

type FormResponse struct {
	ID          int32     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	UserID      int32     `json:"user_id"`
	Status      string    `json:"status"`
	ShareURL    *string   `json:"share_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func formToResponse(form *domain.Form) FormResponse {
	return FormResponse{
		ID:          form.ID,
		Name:        form.Name,
		Description: form.Description,
		UserID:      form.UserID,
		Status:      string(form.Status),
		ShareURL:    form.ShareURL,
		CreatedAt:   form.CreatedAt,
		UpdatedAt:   form.UpdatedAt,
	}
}

func (h *Handler) CreateForm(c *echo.Context) error {
	var req CreateFormRequest
	if err := c.Bind(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "Invalid request body")
	}

	if req.Name == "" || req.UserID == 0 {
		return respondError(c, http.StatusBadRequest, "Name and user_id are required")
	}

	form := &domain.Form{
		Name:        req.Name,
		Description: req.Description,
		UserID:      req.UserID,
	}

	if err := h.formService.CreateForm(c.Request().Context(), form); err != nil {
		return respondError(c, http.StatusInternalServerError, "Failed to create form")
	}

	return c.JSON(http.StatusCreated, formToResponse(form))
}

func (h *Handler) GetForm(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return respondError(c, http.StatusBadRequest, "Invalid form ID")
	}

	form, err := h.formService.GetFormByID(c.Request().Context(), int32(id))
	if err != nil {
		return respondError(c, http.StatusNotFound, "Form not found")
	}

	return c.JSON(http.StatusOK, formToResponse(form))
}

func (h *Handler) GetUserForms(c *echo.Context) error {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return respondError(c, http.StatusBadRequest, "Invalid user ID")
	}

	forms, err := h.formService.GetUserForms(c.Request().Context(), int32(userID))
	if err != nil {
		return respondError(c, http.StatusInternalServerError, "Failed to fetch forms")
	}

	response := make([]FormResponse, len(forms))
	for i, form := range forms {
		response[i] = formToResponse(form)
	}

	return c.JSON(http.StatusOK, response)
}

func (h *Handler) PublishForm(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return respondError(c, http.StatusBadRequest, "Invalid form ID")
	}

	form, err := h.formService.PublishForm(c.Request().Context(), int32(id))
	if err != nil {
		return respondError(c, http.StatusInternalServerError, "Failed to publish form")
	}

	return c.JSON(http.StatusOK, formToResponse(form))
}

func (h *Handler) UnpublishForm(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return respondError(c, http.StatusBadRequest, "Invalid form ID")
	}

	form, err := h.formService.UnpublishForm(c.Request().Context(), int32(id))
	if err != nil {
		return respondError(c, http.StatusInternalServerError, "Failed to unpublish form")
	}

	return c.JSON(http.StatusOK, formToResponse(form))
}

func (h *Handler) DeleteForm(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return respondError(c, http.StatusBadRequest, "Invalid form ID")
	}

	if err := h.formService.DeleteForm(c.Request().Context(), int32(id)); err != nil {
		return respondError(c, http.StatusNotFound, "Form not found")
	}

	return c.NoContent(http.StatusNoContent)
}
