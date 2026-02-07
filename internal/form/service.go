package form

import (
	"context"

	"formify/server/internal/shared"
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
	form, err := s.formRepo.UpdateStatus(ctx, id, StatusPublished)
	if err != nil {
		return nil, err
	}
	if form.ShareURL == nil {
		shareURL, err := shared.GenerateShareURL(12)
		if err != nil {
			return nil, err
		}
		return s.formRepo.UpdateShareURL(ctx, id, shareURL)
	}
	return form, nil
}

func (s *Service) UnpublishForm(ctx context.Context, id int32) (*Form, error) {
	return s.formRepo.UpdateStatus(ctx, id, StatusDraft)
}

func (s *Service) SetShareURL(ctx context.Context, id int32, shareURL string) (*Form, error) {
	return s.formRepo.UpdateShareURL(ctx, id, shareURL)
}

func (s *Service) DeleteForm(ctx context.Context, id int32) error {
	return s.formRepo.Delete(ctx, id)
}

func (s *Service) GetFormOwnerID(ctx context.Context, formID int32) (int32, error) {
	form, err := s.formRepo.GetByID(ctx, formID)
	if err != nil {
		return 0, err
	}
	return form.UserID, nil
}

func (s *Service) IsPublished(ctx context.Context, formID int32) (bool, error) {
	form, err := s.formRepo.GetByID(ctx, formID)
	if err != nil {
		return false, err
	}
	return form.Status == StatusPublished, nil
}
