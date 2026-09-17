package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestHandler_GoogleLogin_SetsProviderAndRedirects(t *testing.T) {
	InitProviders(
		"client-id",
		"client-secret",
		"http://localhost:1323/auth/google/callback",
		"session-secret",
	)

	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/auth/google", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewHandler(nil, nil, "http://localhost:1323", false)
	if err := h.GoogleLogin(c); err != nil {
		t.Fatalf("GoogleLogin returned error: %v", err)
	}

	if got := c.Request().URL.Query().Get("provider"); got != googleProvider {
		t.Fatalf("expected provider query param %q, got %q", googleProvider, got)
	}

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, rec.Code)
	}

	location := rec.Header().Get(echo.HeaderLocation)
	if location == "" {
		t.Fatal("expected redirect location header to be set")
	}
	if !strings.Contains(location, "accounts.google.com") {
		t.Fatalf("expected google auth redirect, got %q", location)
	}
}

func TestHandler_GoogleCallback_UnauthorizedOnOAuthFailure(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/auth/google/callback", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewHandler(nil, nil, "http://localhost:1323", false)
	if err := h.GoogleCallback(c); err != nil {
		t.Fatalf("GoogleCallback returned error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	errMsg := body["error"]
	if !strings.Contains(errMsg, "OAuth authentication failed") {
		t.Fatalf("expected OAuth failure message, got %q", errMsg)
	}
}
