package response

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
)

var errFormNotFound = errors.New("form not found")

type mockFormChecker struct {
	isPublishedFunc    func(ctx context.Context, formID int32) (bool, error)
	getFormOwnerIDFunc func(ctx context.Context, formID int32) (int32, error)
}

func (m *mockFormChecker) IsPublished(ctx context.Context, formID int32) (bool, error) {
	if m.isPublishedFunc != nil {
		return m.isPublishedFunc(ctx, formID)
	}
	return true, nil
}

func (m *mockFormChecker) GetFormOwnerID(ctx context.Context, formID int32) (int32, error) {
	if m.getFormOwnerIDFunc != nil {
		return m.getFormOwnerIDFunc(ctx, formID)
	}
	return 1, nil
}

func TestHandler_CreateResponse(t *testing.T) {
	e := echo.New()

	t.Run("success", func(t *testing.T) {
		mockRepo := &mockRepository{
			createFunc: func(_ context.Context, r *Response) error {
				r.ID = 50
				r.CreatedAt = time.Now()
				return nil
			},
		}
		mockChecker := &mockFormChecker{
			isPublishedFunc: func(_ context.Context, _ int32) (bool, error) {
				return true, nil
			},
		}

		h := NewHandler(NewService(mockRepo, nil, nil, nil), mockChecker)

		body := `{"data":{"q1":"test answer"},"meta":{"ip":"127.0.0.1"}}`
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/forms/5/responses", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/forms/:form_id/responses")
		c.SetPathValues(echo.PathValues{{Name: "form_id", Value: "5"}})

		if err := h.CreateResponse(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", rec.Code)
		}

		var resp ResponseResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.ID != 50 || resp.FormID != 5 {
			t.Fatalf("unexpected response: %+v", resp)
		}
	})

	t.Run("form not published", func(t *testing.T) {
		mockChecker := &mockFormChecker{
			isPublishedFunc: func(_ context.Context, _ int32) (bool, error) {
				return false, nil
			},
		}
		h := NewHandler(NewService(&mockRepository{}, nil, nil, nil), mockChecker)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/forms/5/responses", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/forms/:form_id/responses")
		c.SetPathValues(echo.PathValues{{Name: "form_id", Value: "5"}})

		if err := h.CreateResponse(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", rec.Code)
		}
	})

	t.Run("form not found", func(t *testing.T) {
		mockChecker := &mockFormChecker{
			isPublishedFunc: func(_ context.Context, _ int32) (bool, error) {
				return false, errFormNotFound
			},
		}
		h := NewHandler(NewService(&mockRepository{}, nil, nil, nil), mockChecker)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/forms/5/responses", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/forms/:form_id/responses")
		c.SetPathValues(echo.PathValues{{Name: "form_id", Value: "5"}})

		if err := h.CreateResponse(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", rec.Code)
		}
	})
}

func TestHandler_GetResponse(t *testing.T) {
	e := echo.New()

	t.Run("success", func(t *testing.T) {
		mockRepo := &mockRepository{
			getByIDFunc: func(_ context.Context, id int32) (*Response, error) {
				return &Response{ID: id, FormID: 10}, nil
			},
		}
		mockChecker := &mockFormChecker{
			getFormOwnerIDFunc: func(_ context.Context, _ int32) (int32, error) {
				return 1, nil // owner is user 1
			},
		}
		h := NewHandler(NewService(mockRepo, nil, nil, nil), mockChecker)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/responses/20", http.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/responses/:id")
		c.SetPathValues(echo.PathValues{{Name: "id", Value: "20"}})
		c.Set("user_id", float64(1))

		if err := h.GetResponse(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("forbidden if not owner", func(t *testing.T) {
		mockRepo := &mockRepository{
			getByIDFunc: func(_ context.Context, id int32) (*Response, error) {
				return &Response{ID: id, FormID: 10}, nil
			},
		}
		mockChecker := &mockFormChecker{
			getFormOwnerIDFunc: func(_ context.Context, _ int32) (int32, error) {
				return 99, nil // owner is user 99
			},
		}
		h := NewHandler(NewService(mockRepo, nil, nil, nil), mockChecker)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/responses/20", http.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/responses/:id")
		c.SetPathValues(echo.PathValues{{Name: "id", Value: "20"}})
		c.Set("user_id", float64(1)) // authenticated as user 1

		if err := h.GetResponse(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", rec.Code)
		}
	})
}

func TestHandler_GetFormResponses(t *testing.T) {
	e := echo.New()

	t.Run("success", func(t *testing.T) {
		mockRepo := &mockRepository{
			getByFormIDFunc: func(_ context.Context, formID int32) ([]*Response, error) {
				return []*Response{
					{ID: 1, FormID: formID},
					{ID: 2, FormID: formID},
				}, nil
			},
		}
		mockChecker := &mockFormChecker{
			getFormOwnerIDFunc: func(_ context.Context, _ int32) (int32, error) {
				return 1, nil
			},
		}
		h := NewHandler(NewService(mockRepo, nil, nil, nil), mockChecker)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/forms/10/responses", http.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/forms/:id/responses")
		c.SetPathValues(echo.PathValues{{Name: "id", Value: "10"}})
		c.Set("user_id", float64(1))

		if err := h.GetFormResponses(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp["count"] != float64(2) {
			t.Fatalf("expected count 2, got %v", resp["count"])
		}
	})
}

func TestHandler_DeleteResponse(t *testing.T) {
	e := echo.New()

	mockRepo := &mockRepository{
		getByIDFunc: func(_ context.Context, id int32) (*Response, error) {
			return &Response{ID: id, FormID: 10}, nil
		},
		deleteFunc: func(_ context.Context, _ int32) error {
			return nil
		},
	}
	mockChecker := &mockFormChecker{
		getFormOwnerIDFunc: func(_ context.Context, _ int32) (int32, error) {
			return 1, nil
		},
	}
	h := NewHandler(NewService(mockRepo, nil, nil, nil), mockChecker)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodDelete, "/api/responses/20", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/responses/:id")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "20"}})
	c.Set("user_id", float64(1))

	if err := h.DeleteResponse(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
}
