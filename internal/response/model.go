package response

import (
	"context"
	"time"
)

type Response struct {
	ID        int32     `json:"id"`
	FormID    int32     `json:"form_id"`
	Data      []byte    `json:"data"`
	Meta      []byte    `json:"meta"`
	CreatedAt time.Time `json:"created_at"`
}

// Repository defines response data access methods
type Repository interface {
	Create(ctx context.Context, response *Response) error
	GetByID(ctx context.Context, id int32) (*Response, error)
	GetByFormID(ctx context.Context, formID int32) ([]*Response, error)
	GetByFormIDPaginated(ctx context.Context, formID int32, limit, offset int32) ([]*Response, error)
}
