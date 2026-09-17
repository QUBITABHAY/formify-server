package auth

import (
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

const sessionMaxAge = 86400 // 1 day in seconds

func InitProviders(googleClientID, googleClientSecret, googleCallbackURL, sessionSecret string) {
	store := sessions.NewCookieStore([]byte(sessionSecret))
	store.MaxAge(sessionMaxAge * 1)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	gothic.Store = store

	googleProvider := google.New(
		googleClientID,
		googleClientSecret,
		googleCallbackURL,
		"email",
		"profile",
		"https://www.googleapis.com/auth/drive.file",
		// "https://www.googleapis.com/auth/spreadsheets",
		// "https://www.googleapis.com/auth/drive",
	)
	// Request offline access to get refresh token
	googleProvider.SetAccessType("offline")
	googleProvider.SetPrompt("consent")

	goth.UseProviders(googleProvider)
}
