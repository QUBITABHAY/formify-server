package response

import (
	"context"
	"errors"

	"formify/server/internal/db"
	"formify/server/internal/shared"

	"github.com/jackc/pgx/v5"
)

var ErrResponseNotFound = errors.New("response not found")

type repository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) Repository {
	return &repository{queries: queries}
}

func (r *repository) Create(ctx context.Context, response *Response) error {
	dbResponse, err := r.queries.CreateResponse(ctx, db.CreateResponseParams{
		FormID: response.FormID,
		Data:   response.Data,
		Meta:   response.Meta,
	})
	if err != nil {
		return err
	}
	r.mapDBResponseToModel(dbResponse, response)
	return nil
}

func (r *repository) GetByID(ctx context.Context, id int32) (*Response, error) {
	dbResponse, err := r.queries.GetResponseByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResponseNotFound
		}
		return nil, err
	}
	response := &Response{}
	r.mapDBResponseToModel(dbResponse, response)
	return response, nil
}

func (r *repository) GetByFormID(ctx context.Context, formID int32) ([]*Response, error) {
	dbResponses, err := r.queries.ListResponsesByFormID(ctx, formID)
	if err != nil {
		return nil, err
	}
	return r.mapDBResponsesToModel(dbResponses), nil
}

func (r *repository) GetByFormIDPaginated(ctx context.Context, formID int32, limit, offset int32) ([]*Response, error) {
	dbResponses, err := r.queries.ListResponsesByFormIDPaginated(ctx, db.ListResponsesByFormIDPaginatedParams{
		FormID: formID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	return r.mapDBResponsesToModel(dbResponses), nil
}

func (r *repository) Delete(ctx context.Context, id int32) error {
	return r.queries.DeleteResponse(ctx, id)
}

func (r *repository) DeleteByFormID(ctx context.Context, formID int32) error {
	return r.queries.DeleteResponsesByFormID(ctx, formID)
}

func (r *repository) mapDBResponseToModel(dbResponse db.Response, response *Response) {
	response.ID = dbResponse.ID
	response.FormID = dbResponse.FormID
	response.Data = dbResponse.Data
	response.Meta = dbResponse.Meta
	response.CreatedAt = shared.PgtypeTimestamptzToTime(dbResponse.CreatedAt)
}

func (r *repository) mapDBResponsesToModel(dbResponses []db.Response) []*Response {
	responses := make([]*Response, len(dbResponses))
	for i, dbResponse := range dbResponses {
		responses[i] = &Response{}
		r.mapDBResponseToModel(dbResponse, responses[i])
	}
	return responses
}
