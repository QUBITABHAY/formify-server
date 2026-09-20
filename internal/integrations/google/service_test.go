package google

import (
	"context"
	"errors"
	"testing"
)

func TestInitSheetsService(t *testing.T) {
	t.Run("returns nil when both path and json are empty", func(t *testing.T) {
		svc := InitSheetsService("", "")
		if svc != nil {
			t.Fatal("expected nil service when credentials are empty")
		}
	})

	t.Run("returns nil and logs warning when credentials invalid", func(t *testing.T) {
		svc := InitSheetsService("/non/existent/path.json", "")
		if svc != nil {
			t.Fatal("expected nil service for nonexistent path")
		}
	})
}

func TestNewSheetsService(t *testing.T) {
	t.Run("returns errMissingServiceAccountCreds when both empty", func(t *testing.T) {
		svc, err := NewSheetsService(context.Background(), "", "")
		if svc != nil {
			t.Fatal("expected nil service")
		}
		if !errors.Is(err, errMissingServiceAccountCreds) {
			t.Fatalf("expected errMissingServiceAccountCreds, got %v", err)
		}
	})

	t.Run("returns error when credentialsJSON is invalid", func(t *testing.T) {
		svc, err := NewSheetsService(context.Background(), "", "invalid json")
		if svc != nil {
			t.Fatal("expected nil service")
		}
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
