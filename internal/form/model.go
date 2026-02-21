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
	ID                  int32      `json:"id"`
	FormID              *string    `json:"form_id,omitempty"`
	Name                string     `json:"name"`
	Description         *string    `json:"description,omitempty"`
	UserID              int32      `json:"user_id"`
	Status              Status     `json:"status"`
	Schema              []byte     `json:"schema"`
	Settings            []byte     `json:"settings"`
	ShareURL            *string    `json:"share_url,omitempty"`
	GoogleSheetID       *string    `json:"google_sheet_id,omitempty"`
	GoogleSheetName     *string    `json:"google_sheet_name,omitempty"`
	GoogleSheetLinkedAt *time.Time `json:"google_sheet_linked_at,omitempty"`
	GoogleSheetAutoSync bool       `json:"google_sheet_auto_sync"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type Repository interface {
	Create(ctx context.Context, form *Form) error
	GetByID(ctx context.Context, id int32) (*Form, error)
	GetByShareURL(ctx context.Context, shareURL string) (*Form, error)
	GetByUserID(ctx context.Context, userID int32) ([]*Form, error)
	GetPublishedByUserID(ctx context.Context, userID int32) ([]*Form, error)
	Update(ctx context.Context, form *Form) error
	UpdateStatus(ctx context.Context, id int32, status Status) (*Form, error)
	UpdateShareURL(ctx context.Context, id int32, shareURL string) (*Form, error)
	Delete(ctx context.Context, id int32) error
	LinkGoogleSheet(ctx context.Context, id int32, sheetID, sheetName string, autoSync bool) (*Form, error)
	UnlinkGoogleSheet(ctx context.Context, id int32) (*Form, error)
}
