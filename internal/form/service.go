package form

import (
	"context"
)

type Service struct {
	formRepo Repository
}

func NewService(formRepo Repository) *Service {
	return &Service{formRepo: formRepo}
}

func (s *Service) CreateForm(ctx context.Context, form *Form) error {
	if form.Status == "" {
		form.Status = StatusDraft
	}
	if form.Schema == nil {
		form.Schema = []byte("[]")
	}
	if form.Settings == nil {
		form.Settings = []byte("{}")
	}
	return s.formRepo.Create(ctx, form)
}

func (s *Service) GetFormByID(ctx context.Context, id int32) (*Form, error) {
	return s.formRepo.GetByID(ctx, id)
}

func (s *Service) GetUserForms(ctx context.Context, userID int32) ([]*Form, error) {
	return s.formRepo.GetByUserID(ctx, userID)
}
