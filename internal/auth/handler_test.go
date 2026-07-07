package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestHandler_getCookieDomain(t *testing.T) {
	tests := []struct {
		name        string
		frontendURL string
		want        string
	}{
		{name: "localhost", frontendURL: "http://localhost:1323", want: ""},
		{name: "loopback", frontendURL: "http://127.0.0.1:5173", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(nil, nil, tt.frontendURL, true)
			got := h.getCookieDomain()
			if got != tt.want {
				t.Fatalf("getCookieDomain() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler_getSameSite(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewHandler(nil, nil, "http://localhost:5173", true)
	h.setTokenCookie(c, "abc123")

	res := rec.Result()
	t.Cleanup(func() {
		if err := res.Body.Close(); err != nil {
			t.Errorf("failed to close response body: %v", err)
		}
	})

	var found *http.Cookie
	for _, cookie := range res.Cookies() {
		if cookie.Name == CookieName {
			found = cookie
			break
		}
	}
	if found == nil {
		t.Fatal("cookie not found")
	}
	if found.Value != "abc123" {
		t.Fatalf("cookie value was %v; want abc123", found.Value)
	}
	if found.Path != "/" {
		t.Fatalf("unexpected cookie path: %q", found.Path)
	}
	if !found.HttpOnly {
		t.Fatal("cookie should be HttpOnly")
	}
	if found.SameSite != http.SameSiteNoneMode {
		t.Fatalf("unexpected SameSite: got %v", found.SameSite)
	}
	if !found.Secure {
		t.Fatal("cookie should be Secure when SameSite=None")
	}
}

func TestHandler_clearTokenCookie(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewHandler(nil, nil, "http://localhost:1323", false)
	h.clearTokenCookie(c)

	res := rec.Result()
	t.Cleanup(func() {
		if err := res.Body.Close(); err != nil {
			t.Errorf("failed to close response body: %v", err)
		}
	})

	var found *http.Cookie
	for _, cookie := range res.Cookies() {
		if cookie.Name == CookieName {
			found = cookie
			break
		}
	}
	if found == nil {
		t.Fatal("expected token cookie to be set")
	}
	if found.MaxAge != -1 {
		t.Fatalf("expected cleared MaxAge=-1, got %d", found.MaxAge)
	}
	if found.Value != "" {
		t.Fatalf("expected cleared cookie value, got %q", found.Value)
	}
	if found.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected SameSiteLaxMode, got %v", found.SameSite)
	}
}

func TestHandler_Logout(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewHandler(nil, nil, "http://localhost:1323", false)

	if err := h.Logout(c); err != nil {
		t.Fatalf("Logout returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if body["message"] != "Logged out successfully" {
		t.Fatalf("unexpected message: %q", body["message"])
	}
}

func TestHandler_Me_UnauthorizedWhenUserIDMissing(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/me", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewHandler(nil, nil, "http://localhost:1323", false)

	if err := h.Me(c); err != nil {
		t.Fatalf("Me returned error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if body["error"] != "Unauthorized" {
		t.Fatalf("unexpected error message: %q", body["error"])
	}
}
