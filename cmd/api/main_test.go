package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"

	"formify/server/internal/config"
	"formify/server/internal/logger"
)

func TestSetupMiddlewareAndRoutes(t *testing.T) {
	_ = logger.Init("development")
	e := echo.New()

	cfg := &config.Config{
		Port:                "1323",
		Env:                 "development",
		FrontendURL:         "http://localhost:5173",
		CORSOrigins:         "http://localhost:5173",
		JWTSecret:           "secret",
		SessionSecret:       "session-secret",
		CloudinaryCloudName: "test",
		CloudinaryAPIKey:    "test",
		CloudinaryAPISecret: "test",
	}

	setupMiddleware(e, cfg)
	err := setupRoutes(e, cfg)
	if err != nil {
		t.Fatalf("setupRoutes returned error: %v", err)
	}

	// Test root route
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "Server is running" {
		t.Fatalf("expected 'Server is running', got %q", rec.Body.String())
	}

	// Test health route
	reqHealth := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/health", http.NoBody)
	recHealth := httptest.NewRecorder()
	e.ServeHTTP(recHealth, reqHealth)

	if recHealth.Code != http.StatusOK {
		t.Fatalf("expected status 200 for /health, got %d", recHealth.Code)
	}
}
