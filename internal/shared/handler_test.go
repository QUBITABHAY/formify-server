package shared

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"

	"formify/server/internal/database"
)

func TestRespondError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := RespondError(c, http.StatusBadRequest, "something went wrong")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["error"] != "something went wrong" {
		t.Fatalf("expected error message %q, got %q", "something went wrong", body["error"])
	}
}

func TestHealthCheck(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/health", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := HealthCheck(c); err != nil {
		t.Fatalf("HealthCheck returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	if body["status"] != "ok" {
		t.Fatalf("expected status 'ok', got %q", body["status"])
	}
}

func TestHealthCheckDB_DisconnectedWhenNil(t *testing.T) {
	originalPool := database.DBPool
	database.DBPool = nil
	t.Cleanup(func() {
		database.DBPool = originalPool
	})

	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/health/db", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := HealthCheckDB(c); err != nil {
		t.Fatalf("HealthCheckDB returned error: %v", err)
	}

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	if body["status"] != "db_disconnected" {
		t.Fatalf("expected status 'db_disconnected', got %q", body["status"])
	}
}
