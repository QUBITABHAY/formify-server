// Package response contains response domain handlers, services, and storage.
package response

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"

	"formify/server/internal/shared"
)

type FormChecker interface {
	IsPublished(ctx context.Context, formID int32) (bool, error)
	GetFormOwnerID(ctx context.Context, formID int32) (int32, error)
}

type Handler struct {
	service     *Service
	formChecker FormChecker
}

func NewHandler(service *Service, formChecker FormChecker) *Handler {
	return &Handler{service: service, formChecker: formChecker}
}

type CreateResponseRequest struct {
	Data json.RawMessage `json:"data"`
	Meta json.RawMessage `json:"meta"`
}

type ResponseResponse struct {
	ID        int32           `json:"id"`
	FormID    int32           `json:"form_id"`
	Data      json.RawMessage `json:"data"`
	Meta      json.RawMessage `json:"meta"`
	CreatedAt time.Time       `json:"created_at"`
}

func responseToResponse(resp *Response) ResponseResponse {
	return ResponseResponse{
		ID:        resp.ID,
		FormID:    resp.FormID,
		Data:      resp.Data,
		Meta:      resp.Meta,
		CreatedAt: resp.CreatedAt,
	}
}

func (h *Handler) CreateResponse(c *echo.Context) error {
	formID, err := strconv.ParseInt(c.Param("form_id"), 10, 32)
	if err != nil {
		return shared.RespondError(c, http.StatusBadRequest, "Invalid form ID")
	}

	published, err := h.formChecker.IsPublished(c.Request().Context(), int32(formID))
	if err != nil {
		return shared.RespondError(c, http.StatusNotFound, "Form not found")
	}
	if !published {
		return shared.RespondError(c, http.StatusForbidden, "Form is not accepting responses")
	}

	var req CreateResponseRequest
	if err := c.Bind(&req); err != nil {
		return shared.RespondError(c, http.StatusBadRequest, "Invalid request body")
	}

	response := &Response{
		FormID: int32(formID),
		Data:   req.Data,
		Meta:   req.Meta,
	}

	if err := h.service.CreateResponse(c.Request().Context(), response); err != nil {
		return shared.RespondError(c, http.StatusInternalServerError, "Failed to create response")
	}

	return c.JSON(http.StatusCreated, responseToResponse(response))
}

func (h *Handler) GetResponse(c *echo.Context) error {
	authUserID, ok := shared.GetAuthUserID(c)
	if !ok {
		return shared.RespondError(c, http.StatusUnauthorized, "Unauthorized")
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return shared.RespondError(c, http.StatusBadRequest, "Invalid response ID")
	}

	response, err := h.service.GetResponseByID(c.Request().Context(), int32(id))
	if err != nil {
		return shared.RespondError(c, http.StatusNotFound, "Response not found")
	}

	ownerID, err := h.formChecker.GetFormOwnerID(c.Request().Context(), response.FormID)
	if err != nil {
		return shared.RespondError(c, http.StatusNotFound, "Form not found")
	}
	if ownerID != authUserID {
		return shared.RespondError(c, http.StatusForbidden, "Access denied")
	}

	return c.JSON(http.StatusOK, responseToResponse(response))
}

func (h *Handler) GetFormResponses(c *echo.Context) error {
	authUserID, ok := shared.GetAuthUserID(c)
	if !ok {
		return shared.RespondError(c, http.StatusUnauthorized, "Unauthorized")
	}

	formID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return shared.RespondError(c, http.StatusBadRequest, "Invalid form ID")
	}

	ownerID, err := h.formChecker.GetFormOwnerID(c.Request().Context(), int32(formID))
	if err != nil {
		return shared.RespondError(c, http.StatusNotFound, "Form not found")
	}
	if ownerID != authUserID {
		return shared.RespondError(c, http.StatusForbidden, "Access denied")
	}

	ctx := c.Request().Context()
	responses, err := h.service.GetFormResponses(ctx, int32(formID))
	if err != nil {
		return shared.RespondError(c, http.StatusInternalServerError, "Failed to fetch responses")
	}

	result := make([]ResponseResponse, len(responses))
	for i, resp := range responses {
		result[i] = responseToResponse(resp)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"form_id":   formID,
		"count":     len(responses),
		"responses": result,
	})
}

func (h *Handler) DeleteResponse(c *echo.Context) error {
	authUserID, ok := shared.GetAuthUserID(c)
	if !ok {
		return shared.RespondError(c, http.StatusUnauthorized, "Unauthorized")
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return shared.RespondError(c, http.StatusBadRequest, "Invalid response ID")
	}

	response, err := h.service.GetResponseByID(c.Request().Context(), int32(id))
	if err != nil {
		return shared.RespondError(c, http.StatusNotFound, "Response not found")
	}

	ownerID, err := h.formChecker.GetFormOwnerID(c.Request().Context(), response.FormID)
	if err != nil {
		return shared.RespondError(c, http.StatusNotFound, "Form not found")
	}
	if ownerID != authUserID {
		return shared.RespondError(c, http.StatusForbidden, "Access denied")
	}

	if err := h.service.DeleteResponse(c.Request().Context(), int32(id)); err != nil {
		return shared.RespondError(c, http.StatusInternalServerError, "Failed to delete response")
	}

	return c.NoContent(http.StatusNoContent)
}
