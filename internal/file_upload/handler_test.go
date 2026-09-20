package fileupload

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

var errFormNotFound = errors.New("form not found")

type mockFormChecker struct {
	isPublishedFunc   func(ctx context.Context, formID int32) (bool, error)
	getFormSchemaFunc func(ctx context.Context, formID int32) ([]byte, error)
}

func (m *mockFormChecker) IsPublished(ctx context.Context, formID int32) (bool, error) {
	if m.isPublishedFunc != nil {
		return m.isPublishedFunc(ctx, formID)
	}
	return true, nil
}

func (m *mockFormChecker) GetFormSchema(ctx context.Context, formID int32) ([]byte, error) {
	if m.getFormSchemaFunc != nil {
		return m.getFormSchemaFunc(ctx, formID)
	}
	return []byte(`[{"type":"file"}]`), nil
}

func TestHandler_UploadFile_InvalidFormID(t *testing.T) {
	e := echo.New()
	h := NewHandler(nil, &mockFormChecker{})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/forms/abc/upload", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/forms/:form_id/upload")
	c.SetPathValues(echo.PathValues{{Name: "form_id", Value: "abc"}})

	if err := h.UploadFile(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestHandler_UploadFile_FormNotFound(t *testing.T) {
	e := echo.New()
	checker := &mockFormChecker{
		isPublishedFunc: func(_ context.Context, _ int32) (bool, error) {
			return false, errFormNotFound
		},
	}
	h := NewHandler(nil, checker)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/forms/5/upload", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/forms/:form_id/upload")
	c.SetPathValues(echo.PathValues{{Name: "form_id", Value: "5"}})

	if err := h.UploadFile(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestHandler_UploadFile_FormNotPublished(t *testing.T) {
	e := echo.New()
	checker := &mockFormChecker{
		isPublishedFunc: func(_ context.Context, _ int32) (bool, error) {
			return false, nil
		},
	}
	h := NewHandler(nil, checker)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/forms/5/upload", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/forms/:form_id/upload")
	c.SetPathValues(echo.PathValues{{Name: "form_id", Value: "5"}})

	if err := h.UploadFile(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rec.Code)
	}
}

func TestHandler_UploadFile_SchemaDoesNotAcceptUploads(t *testing.T) {
	e := echo.New()
	checker := &mockFormChecker{
		isPublishedFunc: func(_ context.Context, _ int32) (bool, error) {
			return true, nil
		},
		getFormSchemaFunc: func(_ context.Context, _ int32) ([]byte, error) {
			return []byte(`[{"type":"text"},{"type":"number"}]`), nil
		},
	}
	h := NewHandler(nil, checker)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/forms/5/upload", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/forms/:form_id/upload")
	c.SetPathValues(echo.PathValues{{Name: "form_id", Value: "5"}})

	if err := h.UploadFile(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestHandler_UploadFile_MissingFile(t *testing.T) {
	e := echo.New()
	checker := &mockFormChecker{
		isPublishedFunc: func(_ context.Context, _ int32) (bool, error) {
			return true, nil
		},
		getFormSchemaFunc: func(_ context.Context, _ int32) ([]byte, error) {
			return []byte(`[{"type":"file"}]`), nil
		},
	}
	h := NewHandler(nil, checker)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/forms/5/upload", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/forms/:form_id/upload")
	c.SetPathValues(echo.PathValues{{Name: "form_id", Value: "5"}})

	if err := h.UploadFile(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestSchemaHasFileUpload(t *testing.T) {
	tests := []struct {
		name     string
		schema   string
		expected bool
	}{
		{
			name:     "empty schema",
			schema:   "",
			expected: false,
		},
		{
			name:     "invalid json",
			schema:   "not json",
			expected: false,
		},
		{
			name:     "no file upload field",
			schema:   `[{"type": "text"}, {"type": "email"}]`,
			expected: false,
		},
		{
			name:     "array with file field",
			schema:   `[{"type": "file"}]`,
			expected: true,
		},
		{
			name:     "object with fields array containing upload",
			schema:   `{"fields": [{"elementType": "upload"}]}`,
			expected: true,
		},
		{
			name:     "nested component with attachment",
			schema:   `{"sections": [{"components": [{"component": "attachment"}]}]}`,
			expected: true,
		},
		{
			name:     "field with image_upload",
			schema:   `[{"fieldType": "image_upload"}]`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := schemaHasFileUpload([]byte(tt.schema))
			if got != tt.expected {
				t.Fatalf("schemaHasFileUpload() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsAllowedMIMEType(t *testing.T) {
	allowed := []string{
		"image/jpeg", "image/png", "image/gif", "image/webp", "application/pdf", "application/zip",
	}
	for _, mime := range allowed {
		if !isAllowedMIMEType(mime) {
			t.Errorf("expected %s to be allowed", mime)
		}
	}

	blocked := []string{
		"application/x-msdownload", "text/javascript", "application/x-sh", "text/html",
	}
	for _, mime := range blocked {
		if isAllowedMIMEType(mime) {
			t.Errorf("expected %s to be blocked", mime)
		}
	}
}

func TestRandomHex(t *testing.T) {
	hex1, err := randomHex(8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hex1) != 16 { // 8 bytes = 16 hex chars
		t.Fatalf("expected length 16, got %d", len(hex1))
	}

	hex2, err := randomHex(8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hex1 == hex2 {
		t.Fatal("expected different random values")
	}
}
