// Package fileupload provides file upload validation and storage integration.
package fileupload

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"

	"formify/server/internal/shared"
)

type FormChecker interface {
	IsPublished(ctx context.Context, formID int32) (bool, error)
	GetFormSchema(ctx context.Context, formID int32) ([]byte, error)
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

	schema, err := h.formChecker.GetFormSchema(c.Request().Context(), int32(formID))
	if err != nil {
		return shared.RespondError(c, http.StatusNotFound, "Form not found")
	}
	if !schemaHasFileUpload(schema) {
		return shared.RespondError(c, http.StatusBadRequest, "Form does not accept file uploads")
	}

	file, fileHeader, err := c.Request().FormFile("file")
	if err != nil {
		return shared.RespondError(
			c,
			http.StatusBadRequest,
			"Missing or invalid file field (expected multipart/form-data with field name 'file')",
		)
	}
	defer func() {
		_ = file.Close()
	}()

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

func schemaHasFileUpload(schemaBytes []byte) bool {
	if len(schemaBytes) == 0 {
		return false
	}
	var data any
	if err := json.Unmarshal(schemaBytes, &data); err != nil {
		return false
	}
	return inspectNodeForFileUpload(data)
}

func inspectNodeForFileUpload(node any) bool {
	switch v := node.(type) {
	case map[string]any:
		return inspectMapNode(v)
	case []any:
		return inspectSliceNode(v)
	default:
		return false
	}
}

func inspectMapNode(m map[string]any) bool {
	for k, val := range m {
		if isTypeKey(k) && isTypeValueFileUpload(val) {
			return true
		}
		if inspectNodeForFileUpload(val) {
			return true
		}
	}
	return false
}

func isTypeKey(k string) bool {
	lower := strings.ToLower(k)
	return lower == "type" || lower == "fieldtype" || lower == "elementtype" || lower == "component"
}

func isTypeValueFileUpload(val any) bool {
	s, ok := val.(string)
	if !ok {
		return false
	}
	return isFileUploadType(strings.ToLower(strings.TrimSpace(s)))
}

func inspectSliceNode(s []any) bool {
	return slices.ContainsFunc(s, inspectNodeForFileUpload)
}

func isFileUploadType(t string) bool {
	switch t {
	case "file", "file_upload", "fileupload", "file-upload", "upload", "attachment", "image_upload":
		return true
	default:
		return strings.Contains(t, "file") || strings.Contains(t, "upload")
	}
}
