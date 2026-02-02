package response

import (
	"context"
)

type Service struct {
	responseRepo Repository
}

func NewService(responseRepo Repository) *Service {
	return &Service{responseRepo: responseRepo}
}

func (s *Service) CreateResponse(ctx context.Context, response *Response) error {
	return s.responseRepo.Create(ctx, response)
}

func (s *Service) GetResponseByID(ctx context.Context, id int32) (*Response, error) {
	return s.responseRepo.GetByID(ctx, id)
}

func (s *Service) GetFormResponses(ctx context.Context, formID int32) ([]*Response, error) {
	return s.responseRepo.GetByFormID(ctx, formID)
}

func (s *Service) GetFormResponsesPaginated(ctx context.Context, formID int32, limit, offset int32) ([]*Response, error) {
	return s.responseRepo.GetByFormIDPaginated(ctx, formID, limit, offset)
}

func (s *Service) DeleteResponse(ctx context.Context, id int32) error {
	return s.responseRepo.Delete(ctx, id)
}

func (s *Service) DeleteFormResponses(ctx context.Context, formID int32) error {
	return s.responseRepo.DeleteByFormID(ctx, formID)
}
