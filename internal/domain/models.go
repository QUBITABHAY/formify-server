package domain

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	ID            int32     `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Password      string    `json:"-"`
	OAuthProvider *string   `json:"oauth_provider,omitempty"`
	OAuthID       *string   `json:"oauth_id,omitempty"`
	IsOAuth       bool      `json:"is_oauth"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type FormStatus string

const (
	FormStatusDraft     FormStatus = "draft"
	FormStatusPublished FormStatus = "published"
)

type Form struct {
	ID          int32      `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	UserID      int32      `json:"user_id"`
	Status      FormStatus `json:"status"`
	Schema      []byte     `json:"schema"`
	Settings    []byte     `json:"settings"`
	ShareURL    *string    `json:"share_url,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Response struct {
	ID        int32     `json:"id"`
	FormID    int32     `json:"form_id"`
	Data      []byte    `json:"data"`
	Meta      []byte    `json:"meta"`
	CreatedAt time.Time `json:"created_at"`
}

func PgtypeTextToString(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func StringToPgtypeText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func PgtypeTimestamptzToTime(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func PgtypeBoolToBool(b pgtype.Bool) bool {
	if !b.Valid {
		return false
	}
	return b.Bool
}
