package user

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestHandler_CreateUser_Success(t *testing.T) {
	e := echo.New()
	mock := &mockRepository{
		createFunc: func(_ context.Context, u *User) error {
			u.ID = 10
			return nil
		},
	}
	h := NewHandler(NewService(mock))

	body := `{"name":"Bob","email":"bob@example.com","password":"secretpassword"}`
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.CreateUser(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	var resp UserResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.ID != 10 || resp.Name != "Bob" || resp.Email != "bob@example.com" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestHandler_CreateUser_MissingFields(t *testing.T) {
	e := echo.New()
	mock := &mockRepository{}
	h := NewHandler(NewService(mock))

	body := `{"name":"Bob","email":"","password":""}`
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.CreateUser(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestHandler_CreateUser_InvalidJSON(t *testing.T) {
	e := echo.New()
	mock := &mockRepository{}
	h := NewHandler(NewService(mock))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/users", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.CreateUser(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestHandler_CreateUser_ServiceFailure(t *testing.T) {
	e := echo.New()
	mock := &mockRepository{
		createFunc: func(_ context.Context, _ *User) error {
			return errDB
		},
	}
	h := NewHandler(NewService(mock))

	body := `{"name":"Bob","email":"bob@example.com","password":"secretpassword"}`
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.CreateUser(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}

func TestHandler_GetUser_Success(t *testing.T) {
	e := echo.New()
	mock := &mockRepository{
		getByIDFunc: func(_ context.Context, id int32) (*User, error) {
			return &User{ID: id, Name: "Alice", Email: "alice@example.com"}, nil
		},
	}
	h := NewHandler(NewService(mock))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/users/5", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/users/:id")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "5"}})
	c.Set("user_id", float64(5))

	if err := h.GetUser(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp UserResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.ID != 5 || resp.Name != "Alice" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestHandler_GetUser_Unauthorized(t *testing.T) {
	e := echo.New()
	mock := &mockRepository{}
	h := NewHandler(NewService(mock))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/users/5", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/users/:id")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "5"}})

	if err := h.GetUser(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestHandler_GetUser_InvalidID(t *testing.T) {
	e := echo.New()
	mock := &mockRepository{}
	h := NewHandler(NewService(mock))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/users/abc", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/users/:id")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "abc"}})
	c.Set("user_id", float64(5))

	if err := h.GetUser(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestHandler_GetUser_Forbidden(t *testing.T) {
	e := echo.New()
	mock := &mockRepository{}
	h := NewHandler(NewService(mock))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/users/5", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/users/:id")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "5"}})
	c.Set("user_id", float64(99)) // different user

	if err := h.GetUser(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rec.Code)
	}
}

func TestHandler_GetUser_NotFound(t *testing.T) {
	e := echo.New()
	mock := &mockRepository{
		getByIDFunc: func(_ context.Context, _ int32) (*User, error) {
			return nil, ErrUserNotFound
		},
	}
	h := NewHandler(NewService(mock))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/users/5", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/users/:id")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "5"}})
	c.Set("user_id", float64(5))

	if err := h.GetUser(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}
