package form

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"formify/server/internal/shared"

	"github.com/labstack/echo/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type CreateFormRequest struct {
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	UserID      int32           `json:"user_id"`
	Schema      json.RawMessage `json:"schema,omitempty"`
	Settings    json.RawMessage `json:"settings,omitempty"`
}

type UpdateFormRequest struct {
	Name        string          `json:"name,omitempty"`
	Description *string         `json:"description,omitempty"`
	Schema      json.RawMessage `json:"schema,omitempty"`
	Settings    json.RawMessage `json:"settings,omitempty"`
}

type FormResponse struct {
	ID          int32           `json:"id"`
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	UserID      int32           `json:"user_id"`
	Status      string          `json:"status"`
	Schema      json.RawMessage `json:"schema"`
	Settings    json.RawMessage `json:"settings"`
	ShareURL    *string         `json:"share_url,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func formToResponse(form *Form) FormResponse {
	return FormResponse{
		ID:          form.ID,
		Name:        form.Name,
		Description: form.Description,
		UserID:      form.UserID,
		Status:      string(form.Status),
		Schema:      form.Schema,
		Settings:    form.Settings,
		ShareURL:    form.ShareURL,
		CreatedAt:   form.CreatedAt,
		UpdatedAt:   form.UpdatedAt,
	}
}

func (h *Handler) CreateForm(c *echo.Context) error {
	var req CreateFormRequest
	if err := c.Bind(&req); err != nil {
		return shared.RespondError(c, http.StatusBadRequest, "Invalid request body")
	}

	if req.Name == "" || req.UserID == 0 {
		return shared.RespondError(c, http.StatusBadRequest, "Name and user_id are required")
	}

	form := &Form{
		Name:        req.Name,
		Description: req.Description,
		UserID:      req.UserID,
		Schema:      req.Schema,
		Settings:    req.Settings,
	}

	if err := h.service.CreateForm(c.Request().Context(), form); err != nil {
		return shared.RespondError(c, http.StatusInternalServerError, "Failed to create form")
	}

	return c.JSON(http.StatusCreated, formToResponse(form))
}

func (h *Handler) GetForm(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return shared.RespondError(c, http.StatusBadRequest, "Invalid form ID")
	}

	form, err := h.service.GetFormByID(c.Request().Context(), int32(id))
	if err != nil {
		return shared.RespondError(c, http.StatusNotFound, "Form not found")
	}

	return c.JSON(http.StatusOK, formToResponse(form))
}

func (h *Handler) GetUserForms(c *echo.Context) error {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return shared.RespondError(c, http.StatusBadRequest, "Invalid user ID")
	}

	forms, err := h.service.GetUserForms(c.Request().Context(), int32(userID))
	if err != nil {
		return shared.RespondError(c, http.StatusInternalServerError, "Failed to fetch forms")
	}

	response := make([]FormResponse, len(forms))
	for i, form := range forms {
		response[i] = formToResponse(form)
	}

	return c.JSON(http.StatusOK, response)
}

func (h *Handler) UpdateForm(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return shared.RespondError(c, http.StatusBadRequest, "Invalid form ID")
	}

	existingForm, err := h.service.GetFormByID(c.Request().Context(), int32(id))
	if err != nil {
		return shared.RespondError(c, http.StatusNotFound, "Form not found")
	}

	var req UpdateFormRequest
	if err := c.Bind(&req); err != nil {
		return shared.RespondError(c, http.StatusBadRequest, "Invalid request body")
	}

	if req.Name != "" {
		existingForm.Name = req.Name
	}
	if req.Description != nil {
		existingForm.Description = req.Description
	}
	if req.Schema != nil {
		existingForm.Schema = req.Schema
	}
	if req.Settings != nil {
		existingForm.Settings = req.Settings
	}

	if err := h.service.UpdateForm(c.Request().Context(), existingForm); err != nil {
		return shared.RespondError(c, http.StatusInternalServerError, "Failed to update form")
	}

	return c.JSON(http.StatusOK, formToResponse(existingForm))
}
