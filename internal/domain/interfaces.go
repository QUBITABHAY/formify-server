package domain

import "context"

// UserRepository defines user data access methods
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	CreateOAuth(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id int32) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByOAuthID(ctx context.Context, provider, oauthID string) (*User, error)
	Update(ctx context.Context, user *User) error
	UpdatePassword(ctx context.Context, id int32, password string) error
	Delete(ctx context.Context, id int32) error
	List(ctx context.Context) ([]*User, error)
}

// FormRepository defines form data access methods
type FormRepository interface {
	Create(ctx context.Context, form *Form) error
	GetByID(ctx context.Context, id int32) (*Form, error)
	GetByShareURL(ctx context.Context, shareURL string) (*Form, error)
	GetByUserID(ctx context.Context, userID int32) ([]*Form, error)
	GetPublishedByUserID(ctx context.Context, userID int32) ([]*Form, error)
	Update(ctx context.Context, form *Form) error
	UpdateStatus(ctx context.Context, id int32, status FormStatus) (*Form, error)
	UpdateShareURL(ctx context.Context, id int32, shareURL string) (*Form, error)
	Delete(ctx context.Context, id int32) error
	CountByUserID(ctx context.Context, userID int32) (int64, error)
}

// ResponseRepository defines response data access methods
type ResponseRepository interface {
	Create(ctx context.Context, response *Response) error
	GetByID(ctx context.Context, id int32) (*Response, error)
	GetByFormID(ctx context.Context, formID int32) ([]*Response, error)
	GetByFormIDPaginated(ctx context.Context, formID int32, limit, offset int32) ([]*Response, error)
	Delete(ctx context.Context, id int32) error
	DeleteByFormID(ctx context.Context, formID int32) error
	CountByFormID(ctx context.Context, formID int32) (int64, error)
}
