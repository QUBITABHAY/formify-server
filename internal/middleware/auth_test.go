package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
)

func createTestToken(secret string, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestAuth_SuccessWithCookie(t *testing.T) {
	secret := "test-secret"
	token := createTestToken(secret, jwt.MapClaims{
		"user_id": float64(123),
		"email":   "user@example.com",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/protected", http.NoBody)
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: token,
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var reachedNext bool
	next := func(ctx *echo.Context) error {
		reachedNext = true
		if ctx.Get("user_id") != float64(123) {
			t.Fatalf("expected user_id 123, got %v", ctx.Get("user_id"))
		}
		if ctx.Get("email") != "user@example.com" {
			t.Fatalf("expected email user@example.com, got %v", ctx.Get("email"))
		}
		return ctx.NoContent(http.StatusOK)
	}

	mw := Auth(secret)
	handler := mw(next)

	if err := handler(c); err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}

	if !reachedNext {
		t.Fatal("expected next handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestAuth_SuccessWithBearerHeader(t *testing.T) {
	secret := "test-secret"
	token := createTestToken(secret, jwt.MapClaims{
		"user_id": float64(456),
		"email":   "bearer@example.com",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var reachedNext bool
	next := func(ctx *echo.Context) error {
		reachedNext = true
		if ctx.Get("user_id") != float64(456) {
			t.Fatalf("expected user_id 456, got %v", ctx.Get("user_id"))
		}
		return ctx.NoContent(http.StatusOK)
	}

	mw := Auth(secret)
	handler := mw(next)

	if err := handler(c); err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}

	if !reachedNext {
		t.Fatal("expected next handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestAuth_MissingToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/protected", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := Auth("test-secret")
	handler := mw(func(ctx *echo.Context) error {
		return ctx.NoContent(http.StatusOK)
	})

	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}

	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["error"] != "Unauthorized" {
		t.Fatalf("expected Unauthorized error, got %q", body["error"])
	}
}

func TestAuth_InvalidTokenFormat(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/protected", http.NoBody)
	req.Header.Set("Authorization", "TokenWithoutBearerPrefix")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := Auth("test-secret")
	handler := mw(func(ctx *echo.Context) error {
		return ctx.NoContent(http.StatusOK)
	})

	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}

	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["error"] != "Invalid token format" {
		t.Fatalf("expected Invalid token format, got %q", body["error"])
	}
}

func TestAuth_ExpiredToken(t *testing.T) {
	secret := "test-secret"
	token := createTestToken(secret, jwt.MapClaims{
		"user_id": float64(123),
		"email":   "expired@example.com",
		"exp":     time.Now().Add(-time.Hour).Unix(),
	})

	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := Auth(secret)
	handler := mw(func(ctx *echo.Context) error {
		return ctx.NoContent(http.StatusOK)
	})

	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}

	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["error"] != "Invalid or expired token" {
		t.Fatalf("expected Invalid or expired token, got %q", body["error"])
	}
}

func TestAuth_InvalidSignature(t *testing.T) {
	token := createTestToken("wrong-secret", jwt.MapClaims{
		"user_id": float64(123),
		"email":   "test@example.com",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := Auth("actual-secret")
	handler := mw(func(ctx *echo.Context) error {
		return ctx.NoContent(http.StatusOK)
	})

	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestAuth_InvalidSigningMethod(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"user_id": float64(123),
		"email":   "none@example.com",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/protected", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := Auth("test-secret")
	handler := mw(func(ctx *echo.Context) error {
		return ctx.NoContent(http.StatusOK)
	})

	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}
