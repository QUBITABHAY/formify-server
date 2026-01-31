package form

import (
	"context"
	"time"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
)

type Form struct {
	ID          int32     `json:"id"`
	FormID      *string   `json:"form_id,omitempty"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	UserID      int32     `json:"user_id"`
	Status      Status    `json:"status"`
	Schema      []byte    `json:"schema"`
	Settings    []byte    `json:"settings"`
	ShareURL    *string   `json:"share_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Repository interface {
	Create(ctx context.Context, form *Form) error
	GetByID(ctx context.Context, id int32) (*Form, error)
	GetByUserID(ctx context.Context, userID int32) ([]*Form, error)
	Update(ctx context.Context, form *Form) error
}
