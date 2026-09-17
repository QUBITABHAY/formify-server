package auth

import (
	"testing"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

func TestInitProviders_SetsCookieStore(t *testing.T) {
	previousStore := gothic.Store
	t.Cleanup(func() {
		gothic.Store = previousStore
	})

	InitProviders(
		"client-id",
		"client-secret",
		"http://localhost:1323/auth/google/callback",
		"session-secret")

	if gothic.Store == nil {
		t.Fatal("expected gothic.Store to be initialized")
	}

	store, ok := gothic.Store.(*sessions.CookieStore)
	if !ok {
		t.Fatalf("expected gothic.Store to be *sessions.CookieStore, got %T", gothic.Store)
	}

	if store.Options.Path != "/" {
		t.Fatalf("expected Path=/, got %q", store.Options.Path)
	}

	if !store.Options.HttpOnly {
		t.Fatal("expected HttpOnly=true")
	}

	if store.Options.MaxAge != sessionMaxAge {
		t.Fatalf("expected MaxAge=%d, got %d", sessionMaxAge, store.Options.MaxAge)
	}
}

func TestInitProviders_RegistersGoogleProvider(t *testing.T) {
	InitProviders(
		"client-id",
		"client-secret",
		"http://localhost:1323/auth/google/callback",
		"session-secret",
	)

	provider, err := goth.GetProvider("google")
	if err != nil {
		t.Fatalf("expected google provider to be registered: %v", err)
	}

	if provider == nil {
		t.Fatalf("expected no-nil google provider")
	}

	if provider.Name() != "google" {
		t.Fatalf("expected provider name google, got %q", provider.Name())
	}
}
