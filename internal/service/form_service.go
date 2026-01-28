package service

import (
	"context"

	"formify/server/internal/domain"
)

type FormService struct {
	formRepo domain.FormRepository
}

func NewFormService(formRepo domain.FormRepository) *FormService {
	return &FormService{formRepo: formRepo}
}

func (s *FormService) CreateForm(ctx context.Context, form *domain.Form) error {
	if form.Status == "" {
		form.Status = domain.FormStatusDraft
	}
	if form.Schema == nil {
		form.Schema = []byte("[]")
	}
	if form.Settings == nil {
		form.Settings = []byte("{}")
	}
	return s.formRepo.Create(ctx, form)
}

func (s *FormService) GetFormByID(ctx context.Context, id int32) (*domain.Form, error) {
	return s.formRepo.GetByID(ctx, id)
}

func (s *FormService) GetFormByShareURL(ctx context.Context, shareURL string) (*domain.Form, error) {
	return s.formRepo.GetByShareURL(ctx, shareURL)
}

func (s *FormService) GetUserForms(ctx context.Context, userID int32) ([]*domain.Form, error) {
	return s.formRepo.GetByUserID(ctx, userID)
}

func (s *FormService) GetUserPublishedForms(ctx context.Context, userID int32) ([]*domain.Form, error) {
	return s.formRepo.GetPublishedByUserID(ctx, userID)
}

func (s *FormService) UpdateForm(ctx context.Context, form *domain.Form) error {
	return s.formRepo.Update(ctx, form)
}

func (s *FormService) PublishForm(ctx context.Context, id int32) (*domain.Form, error) {
	return s.formRepo.UpdateStatus(ctx, id, domain.FormStatusPublished)
}

func (s *FormService) UnpublishForm(ctx context.Context, id int32) (*domain.Form, error) {
	return s.formRepo.UpdateStatus(ctx, id, domain.FormStatusDraft)
}

func (s *FormService) SetShareURL(ctx context.Context, id int32, shareURL string) (*domain.Form, error) {
	return s.formRepo.UpdateShareURL(ctx, id, shareURL)
}

func (s *FormService) DeleteForm(ctx context.Context, id int32) error {
	return s.formRepo.Delete(ctx, id)
}

func (s *FormService) CountUserForms(ctx context.Context, userID int32) (int64, error) {
	return s.formRepo.CountByUserID(ctx, userID)
}
