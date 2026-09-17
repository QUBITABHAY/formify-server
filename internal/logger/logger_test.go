package logger

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestInit(t *testing.T) {
	t.Run("development logger", func(t *testing.T) {
		err := Init("development")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if Logger == nil || Sugar == nil {
			t.Fatal("expected Logger and Sugar to be initialized")
		}
	})

	t.Run("production logger", func(t *testing.T) {
		err := Init("production")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if Logger == nil || Sugar == nil {
			t.Fatal("expected Logger and Sugar to be initialized")
		}
	})
}

func TestGetLoggerAndSugar(t *testing.T) {
	_ = Init("development")
	if GetLogger() == nil {
		t.Fatal("expected non-nil logger")
	}
	if GetSugaredLogger() == nil {
		t.Fatal("expected non-nil sugared logger")
	}
}

func TestClose(_ *testing.T) {
	_ = Init("development")
	_ = Close()
}

func TestInitFromEnv(t *testing.T) {
	t.Setenv("ENV", "production")
	if err := InitFromEnv(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Setenv("ENV", "")
	if err := InitFromEnv(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestToField(t *testing.T) {
	field := ToField("my_key", "my_val")
	if field.Key != "my_key" || field.String != "my_val" {
		t.Fatalf("unexpected field: %+v", field)
	}
}

func TestToFields(t *testing.T) {
	t.Run("even arguments", func(t *testing.T) {
		fields := ToFields("k1", "v1", "k2", 123)
		if len(fields) != 2 {
			t.Fatalf("expected 2 fields, got %d", len(fields))
		}
		if fields[0].Key != "k1" || fields[1].Key != "k2" {
			t.Fatalf("unexpected fields: %+v", fields)
		}
	})

	t.Run("odd arguments returns error field", func(t *testing.T) {
		fields := ToFields("odd_key")
		if len(fields) != 1 {
			t.Fatalf("expected 1 error field, got %d", len(fields))
		}
		if fields[0].Key != "error" {
			t.Fatalf("expected error field, got key %q", fields[0].Key)
		}
	})
}

func TestRequestLogger(t *testing.T) {
	_ = Init("development")
	e := echo.New()
	mw := RequestLogger()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := mw(func(ctx *echo.Context) error {
		return ctx.String(http.StatusOK, "logged")
	})

	if err := handler(c); err != nil {
		t.Fatalf("unexpected error from RequestLogger: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
