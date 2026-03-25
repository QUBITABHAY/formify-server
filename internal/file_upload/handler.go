package fileupload

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"formify/server/internal/shared"

	"github.com/labstack/echo/v5"
)

type FormChecker interface {
	IsPublished(ctx context.Context, formID int32) (bool, error)
}

type Handler struct {
	service     *Service
	formChecker FormChecker
}

func NewHandler(service *Service, formChecker FormChecker) *Handler {
	return &Handler{service: service, formChecker: formChecker}
}

func (h *Handler) UploadFile(c *echo.Context) error {
	formID, err := strconv.ParseInt(c.Param("form_id"), 10, 32)
	if err != nil {
		return shared.RespondError(c, http.StatusBadRequest, "Invalid form ID")
	}

	published, err := h.formChecker.IsPublished(c.Request().Context(), int32(formID))
	if err != nil {
		return shared.RespondError(c, http.StatusNotFound, "Form not found")
	}
	if !published {
		return shared.RespondError(c, http.StatusForbidden, "Form is not accepting submissions")
	}

	file, fileHeader, err := c.Request().FormFile("file")
	if err != nil {
		return shared.RespondError(
			c,
			http.StatusBadRequest,
			"Missing or invalid file field (expected multipart/form-data with field name 'file')",
		)
	}
	defer file.Close()

	result, err := h.service.UploadFile(c.Request().Context(), strconv.FormatInt(formID, 10), file, fileHeader)
	if err != nil {
		var valErr *ValidationError
		if errors.As(err, &valErr) {
			return shared.RespondError(c, http.StatusBadRequest, valErr.Message)
		}
		return shared.RespondError(c, http.StatusInternalServerError, "File upload failed")
	}

	return c.JSON(http.StatusOK, result)
}
