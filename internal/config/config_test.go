package config

import (
	"testing"
)

func TestConfig_IsProduction(t *testing.T) {
	tests := []struct {
		env  string
		want bool
	}{
		{"production", true},
		{"development", false},
		{"test", false},
		{"", false},
	}

	for _, tt := range tests {
		cfg := &Config{Env: tt.env}
		if got := cfg.IsProduction(); got != tt.want {
			t.Errorf("IsProduction() for env=%q got %v, want %v", tt.env, got, tt.want)
		}
	}
}

func TestConfig_GetCORSOrigins(t *testing.T) {
	t.Run("returns FrontendURL when CORSOrigins is empty", func(t *testing.T) {
		cfg := &Config{
			FrontendURL: "http://localhost:5173",
			CORSOrigins: "",
		}
		got := cfg.GetCORSOrigins()
		if len(got) != 1 || got[0] != "http://localhost:5173" {
			t.Fatalf("expected ['http://localhost:5173'], got %v", got)
		}
	})

	t.Run("returns split origins when CORSOrigins is provided", func(t *testing.T) {
		cfg := &Config{
			FrontendURL: "http://localhost:5173",
			CORSOrigins: "http://localhost:5173,https://example.com,https://app.formify.io",
		}
		got := cfg.GetCORSOrigins()
		expected := []string{"http://localhost:5173", "https://example.com", "https://app.formify.io"}
		if len(got) != len(expected) {
			t.Fatalf("expected %d origins, got %d", len(expected), len(got))
		}
		for i := range expected {
			if got[i] != expected[i] {
				t.Fatalf("at index %d: expected %q, got %q", i, expected[i], got[i])
			}
		}
	})
}

func TestLoad(t *testing.T) {
	t.Setenv("PORT", "9999")
	t.Setenv("ENV", "test-env")
	t.Setenv("FRONTEND_URL", "http://test.local")

	cfg := Load()
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	if cfg.Port != "9999" {
		t.Fatalf("expected Port 9999, got %q", cfg.Port)
	}
	if cfg.Env != "test-env" {
		t.Fatalf("expected Env test-env, got %q", cfg.Env)
	}
}
