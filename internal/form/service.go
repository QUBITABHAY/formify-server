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

func (s *Service) GetFormByShareURL(ctx context.Context, shareURL string) (*Form, error) {
	return s.formRepo.GetByShareURL(ctx, shareURL)
}

func (s *Service) GetUserForms(ctx context.Context, userID int32) ([]*Form, error) {
	return s.formRepo.GetByUserID(ctx, userID)
}

func (s *Service) GetUserPublishedForms(ctx context.Context, userID int32) ([]*Form, error) {
	return s.formRepo.GetPublishedByUserID(ctx, userID)
}

func (s *Service) UpdateForm(ctx context.Context, form *Form) error {
	return s.formRepo.Update(ctx, form)
}

func (s *Service) PublishForm(ctx context.Context, id int32) (*Form, error) {
	return s.formRepo.UpdateStatus(ctx, id, StatusPublished)
}

func (s *Service) UnpublishForm(ctx context.Context, id int32) (*Form, error) {
	return s.formRepo.UpdateStatus(ctx, id, StatusDraft)
}

func (s *Service) SetShareURL(ctx context.Context, id int32, shareURL string) (*Form, error) {
	return s.formRepo.UpdateShareURL(ctx, id, shareURL)
}
