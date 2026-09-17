package form

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestHandler_CreateForm_Success(t *testing.T) {
	e := echo.New()
	mockRepo := &mockFormRepository{
		createFunc: func(_ context.Context, f *Form) error {
			f.ID = 101
			return nil
		},
	}
	h := NewHandler(NewService(mockRepo, nil), nil, nil, nil)

	body := `{"name":"Customer Feedback","description":"Survey"}`
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/forms", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", float64(1))

	if err := h.CreateForm(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	var resp FormResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.ID != 101 || resp.Name != "Customer Feedback" || resp.UserID != 1 {
		t.Fatalf("unexpected form response: %+v", resp)
	}
}

func TestHandler_CreateForm_Unauthorized(t *testing.T) {
	e := echo.New()
	h := NewHandler(NewService(&mockFormRepository{}, nil), nil, nil, nil)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/forms", bytes.NewBufferString(`{"name":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.CreateForm(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestHandler_CreateForm_MissingName(t *testing.T) {
	e := echo.New()
	h := NewHandler(NewService(&mockFormRepository{}, nil), nil, nil, nil)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/forms", bytes.NewBufferString(`{"name":""}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", float64(1))

	if err := h.CreateForm(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestHandler_GetForm(t *testing.T) {
	e := echo.New()

	t.Run("success", func(t *testing.T) {
		mockRepo := &mockFormRepository{
			getByIDFunc: func(_ context.Context, id int32) (*Form, error) {
				return &Form{ID: id, Name: "My Form", UserID: 1}, nil
			},
		}
		h := NewHandler(NewService(mockRepo, nil), nil, nil, nil)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/forms/10", http.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/forms/:id")
		c.SetPathValues(echo.PathValues{{Name: "id", Value: "10"}})
		c.Set("user_id", float64(1))

		if err := h.GetForm(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("forbidden for different user", func(t *testing.T) {
		mockRepo := &mockFormRepository{
			getByIDFunc: func(_ context.Context, id int32) (*Form, error) {
				return &Form{ID: id, Name: "My Form", UserID: 99}, nil
			},
		}
		h := NewHandler(NewService(mockRepo, nil), nil, nil, nil)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/forms/10", http.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/forms/:id")
		c.SetPathValues(echo.PathValues{{Name: "id", Value: "10"}})
		c.Set("user_id", float64(1)) // User 1 trying to access User 99's form

		if err := h.GetForm(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", rec.Code)
		}
	})
}

func TestHandler_GetUserForms(t *testing.T) {
	e := echo.New()

	t.Run("success", func(t *testing.T) {
		mockRepo := &mockFormRepository{
			getByUserIDFunc: func(_ context.Context, userID int32) ([]*Form, error) {
				return []*Form{
					{ID: 1, Name: "Form 1", UserID: userID},
					{ID: 2, Name: "Form 2", UserID: userID},
				}, nil
			},
		}
		h := NewHandler(NewService(mockRepo, nil), nil, nil, nil)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/users/5/forms", http.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/users/:id/forms")
		c.SetPathValues(echo.PathValues{{Name: "id", Value: "5"}})
		c.Set("user_id", float64(5))

		if err := h.GetUserForms(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var forms []FormResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &forms); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(forms) != 2 {
			t.Fatalf("expected 2 forms, got %d", len(forms))
		}
	})
}

func TestHandler_GetPublicFormsByShareURL(t *testing.T) {
	e := echo.New()
	sheetID := "sheet-id"

	t.Run("success published form strips answers and sheet metadata", func(t *testing.T) {
		schema := []byte(`[{"id":"q1","label":"What is 2+2?","correctAnswer":"4"}]`)
		mockRepo := &mockFormRepository{
			getByShareURLFunc: func(_ context.Context, _ string) (*Form, error) {
				return &Form{
					ID:                  1,
					Name:                "Quiz",
					Status:              StatusPublished,
					Schema:              schema,
					GoogleSheetID:       &sheetID,
					GoogleSheetAutoSync: true,
				}, nil
			},
		}
		h := NewHandler(NewService(mockRepo, nil), nil, nil, nil)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/forms/share/token123", http.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/forms/share/:share_url")
		c.SetPathValues(echo.PathValues{{Name: "share_url", Value: "token123"}})

		if err := h.GetPublicFormsByShareURL(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp FormResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.GoogleSheetID != nil || resp.GoogleSheetAutoSync {
			t.Fatal("expected google sheet fields to be stripped")
		}

		var schemaSlice []map[string]any
		if err := json.Unmarshal(resp.Schema, &schemaSlice); err != nil {
			t.Fatalf("failed to unmarshal cleaned schema: %v", err)
		}
		if _, exists := schemaSlice[0]["correctAnswer"]; exists {
			t.Fatal("expected correctAnswer to be stripped from schema")
		}
	})

	t.Run("not found if form is draft", func(t *testing.T) {
		mockRepo := &mockFormRepository{
			getByShareURLFunc: func(_ context.Context, _ string) (*Form, error) {
				return &Form{
					ID:     1,
					Name:   "Draft Quiz",
					Status: StatusDraft,
				}, nil
			},
		}
		h := NewHandler(NewService(mockRepo, nil), nil, nil, nil)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/forms/share/token123", http.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/forms/share/:share_url")
		c.SetPathValues(echo.PathValues{{Name: "share_url", Value: "token123"}})

		if err := h.GetPublicFormsByShareURL(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status 404 for draft form, got %d", rec.Code)
		}
	})
}

func TestHandler_UpdateForm(t *testing.T) {
	e := echo.New()

	t.Run("success", func(t *testing.T) {
		mockRepo := &mockFormRepository{
			getByIDFunc: func(_ context.Context, id int32) (*Form, error) {
				return &Form{ID: id, Name: "Old Name", UserID: 1}, nil
			},
			updateFunc: func(_ context.Context, _ *Form) error {
				return nil
			},
		}
		h := NewHandler(NewService(mockRepo, nil), nil, nil, nil)

		body := `{"name":"Updated Name"}`
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/api/forms/5", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/forms/:id")
		c.SetPathValues(echo.PathValues{{Name: "id", Value: "5"}})
		c.Set("user_id", float64(1))

		if err := h.UpdateForm(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp FormResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp.Name != "Updated Name" {
			t.Fatalf("expected Name 'Updated Name', got %s", resp.Name)
		}
	})
}

func TestHandler_PublishAndUnpublishForm(t *testing.T) {
	e := echo.New()

	mockRepo := &mockFormRepository{
		getByIDFunc: func(_ context.Context, id int32) (*Form, error) {
			return &Form{ID: id, Name: "Form", UserID: 1}, nil
		},
		updateStatusFunc: func(_ context.Context, id int32, status Status) (*Form, error) {
			return &Form{ID: id, Name: "Form", UserID: 1, Status: status}, nil
		},
		updateShareURLFunc: func(_ context.Context, id int32, shareURL string) (*Form, error) {
			return &Form{ID: id, Name: "Form", UserID: 1, Status: StatusPublished, ShareURL: &shareURL}, nil
		},
	}
	h := NewHandler(NewService(mockRepo, nil), nil, nil, nil)

	t.Run("publish", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/forms/5/publish", http.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/forms/:id/publish")
		c.SetPathValues(echo.PathValues{{Name: "id", Value: "5"}})
		c.Set("user_id", float64(1))

		if err := h.PublishForm(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("unpublish", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/forms/5/unpublish", http.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/forms/:id/unpublish")
		c.SetPathValues(echo.PathValues{{Name: "id", Value: "5"}})
		c.Set("user_id", float64(1))

		if err := h.UnpublishForm(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})
}

func TestHandler_DeleteForm(t *testing.T) {
	e := echo.New()

	mockRepo := &mockFormRepository{
		getByIDFunc: func(_ context.Context, id int32) (*Form, error) {
			return &Form{ID: id, Name: "Form", UserID: 1}, nil
		},
		deleteFunc: func(_ context.Context, _ int32) error {
			return nil
		},
	}
	mockResp := &mockResponseRepository{
		deleteByFormIDFunc: func(_ context.Context, _ int32) error {
			return nil
		},
	}
	h := NewHandler(NewService(mockRepo, mockResp), nil, nil, nil)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodDelete, "/api/forms/5", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/forms/:id")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "5"}})
	c.Set("user_id", float64(1))

	if err := h.DeleteForm(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
}

func TestHandler_UnlinkGoogleSheet(t *testing.T) {
	e := echo.New()

	mockRepo := &mockFormRepository{
		getByIDFunc: func(_ context.Context, id int32) (*Form, error) {
			return &Form{ID: id, Name: "Form", UserID: 1}, nil
		},
		unlinkGoogleSheetFunc: func(_ context.Context, id int32) (*Form, error) {
			return &Form{ID: id, Name: "Form", UserID: 1, GoogleSheetID: nil}, nil
		},
	}
	h := NewHandler(NewService(mockRepo, nil), nil, nil, nil)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodDelete, "/api/forms/5/sheets/link", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/forms/:id/sheets/link")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "5"}})
	c.Set("user_id", float64(1))

	if err := h.UnlinkGoogleSheet(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestStripCorrectAnswersFromSchema(t *testing.T) {
	t.Run("removes correctAnswer and correct_answer deeply", func(t *testing.T) {
		input := []byte(`{
			"title": "Quiz",
			"correctAnswer": "top-level",
			"fields": [
				{"id": "q1", "correctAnswer": "A"},
				{"id": "q2", "correct_answer": "B", "nested": {"correctAnswer": "C"}}
			]
		}`)

		cleaned := stripCorrectAnswersFromSchema(input)
		var parsed map[string]any
		if err := json.Unmarshal(cleaned, &parsed); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if _, ok := parsed["correctAnswer"]; ok {
			t.Fatal("expected top-level correctAnswer to be stripped")
		}

		fields := parsed["fields"].([]any)
		f1 := fields[0].(map[string]any)
		if _, ok := f1["correctAnswer"]; ok {
			t.Fatal("expected q1 correctAnswer to be stripped")
		}

		f2 := fields[1].(map[string]any)
		if _, ok := f2["correct_answer"]; ok {
			t.Fatal("expected q2 correct_answer to be stripped")
		}
		nested := f2["nested"].(map[string]any)
		if _, ok := nested["correctAnswer"]; ok {
			t.Fatal("expected nested correctAnswer to be stripped")
		}
	})

	t.Run("handles empty or invalid JSON", func(t *testing.T) {
		if string(stripCorrectAnswersFromSchema(nil)) != "" {
			t.Fatal("expected empty result for nil")
		}
		invalid := []byte("invalid json")
		if string(stripCorrectAnswersFromSchema(invalid)) != "invalid json" {
			t.Fatal("expected invalid json to return unchanged")
		}
	})
}
