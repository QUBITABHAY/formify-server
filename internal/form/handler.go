package form

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"

	"formify/server/internal/integrations/google"
	responsepkg "formify/server/internal/response"
	"formify/server/internal/shared"
	"formify/server/internal/user"
)

type Handler struct {
	service       *Service
	sheetsService *google.SheetsService
	userService   *user.Service
	responseSvc   *responsepkg.Service
}

func NewHandler(
	service *Service,
	sheetsService *google.SheetsService,
	userService *user.Service,
	responseSvc *responsepkg.Service,
) *Handler {
	return &Handler{service: service, sheetsService: sheetsService, userService: userService, responseSvc: responseSvc}
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
	ID                  int32           `json:"id"`
	Name                string          `json:"name"`
	Description         *string         `json:"description,omitempty"`
	UserID              int32           `json:"user_id"`
	Status              string          `json:"status"`
	Schema              json.RawMessage `json:"schema"`
	Settings            json.RawMessage `json:"settings"`
	ShareURL            *string         `json:"share_url,omitempty"`
	GoogleSheetID       *string         `json:"google_sheet_id,omitempty"`
	GoogleSheetName     *string         `json:"google_sheet_name,omitempty"`
	GoogleSheetLinkedAt *time.Time      `json:"google_sheet_linked_at,omitempty"`
	GoogleSheetAutoSync bool            `json:"google_sheet_auto_sync"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

type CreateGoogleSheetRequest struct {
	Title string `json:"title"`
}

func formToResponse(form *Form) FormResponse {
	return FormResponse{
		ID:                  form.ID,
		Name:                form.Name,
		Description:         form.Description,
		UserID:              form.UserID,
		Status:              string(form.Status),
		Schema:              form.Schema,
		Settings:            form.Settings,
		ShareURL:            form.ShareURL,
		GoogleSheetID:       form.GoogleSheetID,
		GoogleSheetName:     form.GoogleSheetName,
		GoogleSheetLinkedAt: form.GoogleSheetLinkedAt,
		GoogleSheetAutoSync: form.GoogleSheetAutoSync,
		CreatedAt:           form.CreatedAt,
		UpdatedAt:           form.UpdatedAt,
	}
}

func (h *Handler) CreateForm(c *echo.Context) error {
	authUserID, ok := shared.GetAuthUserID(c)
	if !ok {
		return shared.RespondError(c, http.StatusUnauthorized, "Unauthorized")
	}

	var req CreateFormRequest
	if err := c.Bind(&req); err != nil {
		return shared.RespondError(c, http.StatusBadRequest, "Invalid request body")
	}

	if req.Name == "" {
		return shared.RespondError(c, http.StatusBadRequest, "Name is required")
	}

	formID := base64.RawURLEncoding.EncodeToString([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)))
	form := &Form{
		FormID:      &formID,
		Name:        req.Name,
		Description: req.Description,
		UserID:      authUserID,
		Schema:      req.Schema,
		Settings:    req.Settings,
	}

	if err := h.service.CreateForm(c.Request().Context(), form); err != nil {
		return shared.RespondError(c, http.StatusInternalServerError, "Failed to create form")
	}

	return c.JSON(http.StatusCreated, formToResponse(form))
}

func (h *Handler) GetForm(c *echo.Context) error {
	authUserID, ok := shared.GetAuthUserID(c)
	if !ok {
		return shared.RespondError(c, http.StatusUnauthorized, "Unauthorized")
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return shared.RespondError(c, http.StatusBadRequest, "Invalid form ID")
	}

	form, err := h.service.GetFormByID(c.Request().Context(), int32(id))
	if err != nil {
		return shared.RespondError(c, http.StatusNotFound, "Form not found")
	}

	if form.UserID != authUserID {
		return shared.RespondError(c, http.StatusForbidden, "Access denied")
	}

	return c.JSON(http.StatusOK, formToResponse(form))
}

func (h *Handler) GetUserForms(c *echo.Context) error {
	authUserID, ok := shared.GetAuthUserID(c)
	if !ok {
		return shared.RespondError(c, http.StatusUnauthorized, "Unauthorized")
	}

	userID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return shared.RespondError(c, http.StatusBadRequest, "Invalid user ID")
	}

	if int32(userID) != authUserID {
		return shared.RespondError(c, http.StatusForbidden, "Access denied")
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

func (h *Handler) GetPublicFormsByShareURL(c *echo.Context) error {
	shareURL := c.Param("share_url")
	if shareURL == "" {
		return shared.RespondError(c, http.StatusBadRequest, "Invalid share URL")
	}

	form, err := h.service.GetFormByShareURL(c.Request().Context(), shareURL)
	if err != nil {
		return shared.RespondError(c, http.StatusNotFound, "Form not found")
	}

	if form.Status != StatusPublished {
		return shared.RespondError(c, http.StatusNotFound, "Form not found")
	}

	return c.JSON(http.StatusOK, formToResponse(form))
}

func (h *Handler) UpdateForm(c *echo.Context) error {
	authUserID, ok := shared.GetAuthUserID(c)
	if !ok {
		return shared.RespondError(c, http.StatusUnauthorized, "Unauthorized")
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return shared.RespondError(c, http.StatusBadRequest, "Invalid form ID")
	}

	existingForm, err := h.service.GetFormByID(c.Request().Context(), int32(id))
	if err != nil {
		return shared.RespondError(c, http.StatusNotFound, "Form not found")
	}

	if existingForm.UserID != authUserID {
		return shared.RespondError(c, http.StatusForbidden, "Access denied")
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

func (h *Handler) getAuthorizedForm(c *echo.Context) (int32, error) {
	authUserID, ok := shared.GetAuthUserID(c)
	if !ok {
		return 0, shared.RespondError(c, http.StatusUnauthorized, "Unauthorized")
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return 0, shared.RespondError(c, http.StatusBadRequest, "Invalid form ID")
	}

	existingForm, err := h.service.GetFormByID(c.Request().Context(), int32(id))
	if err != nil {
		return 0, shared.RespondError(c, http.StatusNotFound, "Form not found")
	}
	if existingForm.UserID != authUserID {
		return 0, shared.RespondError(c, http.StatusForbidden, "Access denied")
	}

	return int32(id), nil
}

func (h *Handler) PublishForm(c *echo.Context) error {
	id, err := h.getAuthorizedForm(c)
	if err != nil {
		return err
	}

	form, err := h.service.PublishForm(c.Request().Context(), id)
	if err != nil {
		return shared.RespondError(c, http.StatusInternalServerError, "Failed to publish form")
	}

	return c.JSON(http.StatusOK, formToResponse(form))
}

func (h *Handler) UnpublishForm(c *echo.Context) error {
	id, err := h.getAuthorizedForm(c)
	if err != nil {
		return err
	}

	form, err := h.service.UnpublishForm(c.Request().Context(), id)
	if err != nil {
		return shared.RespondError(c, http.StatusInternalServerError, "Failed to unpublish form")
	}

	return c.JSON(http.StatusOK, formToResponse(form))
}

func (h *Handler) DeleteForm(c *echo.Context) error {
	authUserID, ok := shared.GetAuthUserID(c)
	if !ok {
		return shared.RespondError(c, http.StatusUnauthorized, "Unauthorized")
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return shared.RespondError(c, http.StatusBadRequest, "Invalid form ID")
	}

	existingForm, err := h.service.GetFormByID(c.Request().Context(), int32(id))
	if err != nil {
		return shared.RespondError(c, http.StatusNotFound, "Form not found")
	}
	if existingForm.UserID != authUserID {
		return shared.RespondError(c, http.StatusForbidden, "Access denied")
	}

	if err := h.service.DeleteForm(c.Request().Context(), int32(id)); err != nil {
		return shared.RespondError(c, http.StatusInternalServerError, "Failed to delete form")
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) CreateAndLinkGoogleSheet(c *echo.Context) error {
	authUserID, ok := shared.GetAuthUserID(c)
	if !ok {
		return shared.RespondError(c, http.StatusUnauthorized, "Unauthorized")
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return shared.RespondError(c, http.StatusBadRequest, "Invalid form ID")
	}

	existingForm, err := h.service.GetFormByID(c.Request().Context(), int32(id))
	if err != nil {
		return shared.RespondError(c, http.StatusNotFound, "Form not found")
	}
	if existingForm.UserID != authUserID {
		return shared.RespondError(c, http.StatusForbidden, "Access denied")
	}

	var req CreateGoogleSheetRequest
	if err := c.Bind(&req); err != nil {
		return shared.RespondError(c, http.StatusBadRequest, "Invalid request body")
	}

	title := req.Title
	if title == "" {
		title = existingForm.Name + " - Responses"
	}

	fields, _ := google.ParseFormSchema(existingForm.Schema)
	headers := google.ExtractHeaders(fields)

	currentUser, err := h.userService.GetUserByID(c.Request().Context(), authUserID)
	if err != nil {
		return shared.RespondError(c, http.StatusInternalServerError, "Failed to get user")
	}

	if currentUser.GoogleAccessToken == nil || *currentUser.GoogleAccessToken == "" {
		return shared.RespondError(c, http.StatusUnauthorized, "Google OAuth is required to link a sheet. Please login with Google.")
	}

	expiry := time.Now()
	if currentUser.GoogleTokenExpiry != nil {
		expiry = *currentUser.GoogleTokenExpiry
	}
	refreshToken := ""
	if currentUser.GoogleRefreshToken != nil {
		refreshToken = *currentUser.GoogleRefreshToken
	}

	sheetsService, err := google.NewSheetsServiceWithUserToken(
		c.Request().Context(),
		*currentUser.GoogleAccessToken,
		refreshToken,
		expiry,
	)
	if err != nil {
		log.Printf("Failed to create user token sheets service: %v", err)
		return shared.RespondError(c, http.StatusUnauthorized, "Google OAuth is required to link a sheet. Please login with Google again.")
	}

	log.Printf("Using user's OAuth token for Google Sheets")

	spreadsheetID, err := sheetsService.CreateSpreadsheet(c.Request().Context(), title, headers)
	if err != nil {
		log.Printf("Failed to create Google Sheet for form %d: %v", id, err)
		return shared.RespondError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create Google Sheet: %v", err))
	}

	if h.responseSvc != nil {
		if err := h.responseSvc.BackfillFormResponsesToSheet(c.Request().Context(), int32(id), existingForm.Schema, spreadsheetID, existingForm.UserID); err != nil {
			log.Printf("Failed to backfill responses for form %d to new sheet %s: %v", id, spreadsheetID, err)
			return shared.RespondError(c, http.StatusInternalServerError, "Failed to sync existing responses to Google Sheet")
		}
	}

	form, err := h.service.LinkGoogleSheet(c.Request().Context(), int32(id), spreadsheetID, title, true)
	if err != nil {
		return shared.RespondError(c, http.StatusInternalServerError, "Failed to link Google Sheet")
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"form":            formToResponse(form),
		"spreadsheet_id":  spreadsheetID,
		"spreadsheet_url": "https://docs.google.com/spreadsheets/d/" + spreadsheetID,
	})
}

func (h *Handler) UnlinkGoogleSheet(c *echo.Context) error {
	id, err := h.getAuthorizedForm(c)
	if err != nil {
		return err
	}

	form, err := h.service.UnlinkGoogleSheet(c.Request().Context(), id)
	if err != nil {
		return shared.RespondError(c, http.StatusInternalServerError, "Failed to unlink Google Sheet")
	}

	return c.JSON(http.StatusOK, formToResponse(form))
}
