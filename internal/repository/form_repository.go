package repository

import (
	"context"
	"errors"

	"formify/server/internal/db"
	"formify/server/internal/domain"

	"github.com/jackc/pgx/v5"
)

var ErrFormNotFound = errors.New("form not found")

type formRepository struct {
	queries *db.Queries
}

func NewFormRepository(queries *db.Queries) domain.FormRepository {
	return &formRepository{queries: queries}
}

func (r *formRepository) Create(ctx context.Context, form *domain.Form) error {
	dbForm, err := r.queries.CreateForm(ctx, db.CreateFormParams{
		Name:        form.Name,
		Description: domain.StringToPgtypeText(form.Description),
		UserID:      form.UserID,
		Status:      db.NullFormStatus{FormStatus: db.FormStatus(form.Status), Valid: form.Status != ""},
		Schema:      form.Schema,
		Settings:    form.Settings,
		ShareUrl:    domain.StringToPgtypeText(form.ShareURL),
	})
	if err != nil {
		return err
	}
	r.mapDBFormToDomain(dbForm, form)
	return nil
}

func (r *formRepository) GetByID(ctx context.Context, id int32) (*domain.Form, error) {
	dbForm, err := r.queries.GetFormByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFormNotFound
		}
		return nil, err
	}
	form := &domain.Form{}
	r.mapDBFormToDomain(dbForm, form)
	return form, nil
}

func (r *formRepository) GetByShareURL(ctx context.Context, shareURL string) (*domain.Form, error) {
	dbForm, err := r.queries.GetFormByShareURL(ctx, domain.StringToPgtypeText(&shareURL))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFormNotFound
		}
		return nil, err
	}
	form := &domain.Form{}
	r.mapDBFormToDomain(dbForm, form)
	return form, nil
}

func (r *formRepository) GetByUserID(ctx context.Context, userID int32) ([]*domain.Form, error) {
	dbForms, err := r.queries.ListFormsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return r.mapDBFormsToDomain(dbForms), nil
}

func (r *formRepository) GetPublishedByUserID(ctx context.Context, userID int32) ([]*domain.Form, error) {
	dbForms, err := r.queries.ListPublishedFormsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return r.mapDBFormsToDomain(dbForms), nil
}

func (r *formRepository) Update(ctx context.Context, form *domain.Form) error {
	dbForm, err := r.queries.UpdateForm(ctx, db.UpdateFormParams{
		ID:          form.ID,
		Name:        form.Name,
		Description: domain.StringToPgtypeText(form.Description),
		Schema:      form.Schema,
		Settings:    form.Settings,
	})
	if err != nil {
		return err
	}
	r.mapDBFormToDomain(dbForm, form)
	return nil
}

func (r *formRepository) UpdateStatus(ctx context.Context, id int32, status domain.FormStatus) (*domain.Form, error) {
	dbForm, err := r.queries.UpdateFormStatus(ctx, db.UpdateFormStatusParams{
		ID:     id,
		Status: db.NullFormStatus{FormStatus: db.FormStatus(status), Valid: true},
	})
	if err != nil {
		return nil, err
	}
	form := &domain.Form{}
	r.mapDBFormToDomain(dbForm, form)
	return form, nil
}

func (r *formRepository) UpdateShareURL(ctx context.Context, id int32, shareURL string) (*domain.Form, error) {
	dbForm, err := r.queries.UpdateFormShareURL(ctx, db.UpdateFormShareURLParams{
		ID:       id,
		ShareUrl: domain.StringToPgtypeText(&shareURL),
	})
	if err != nil {
		return nil, err
	}
	form := &domain.Form{}
	r.mapDBFormToDomain(dbForm, form)
	return form, nil
}

func (r *formRepository) Delete(ctx context.Context, id int32) error {
	return r.queries.DeleteForm(ctx, id)
}

func (r *formRepository) CountByUserID(ctx context.Context, userID int32) (int64, error) {
	return r.queries.CountFormsByUserID(ctx, userID)
}

func (r *formRepository) mapDBFormToDomain(dbForm db.Form, form *domain.Form) {
	form.ID = dbForm.ID
	form.Name = dbForm.Name
	form.Description = domain.PgtypeTextToString(dbForm.Description)
	form.UserID = dbForm.UserID
	if dbForm.Status.Valid {
		form.Status = domain.FormStatus(dbForm.Status.FormStatus)
	}
	form.Schema = dbForm.Schema
	form.Settings = dbForm.Settings
	form.ShareURL = domain.PgtypeTextToString(dbForm.ShareUrl)
	form.CreatedAt = domain.PgtypeTimestamptzToTime(dbForm.CreatedAt)
	form.UpdatedAt = domain.PgtypeTimestamptzToTime(dbForm.UpdatedAt)
}

func (r *formRepository) mapDBFormsToDomain(dbForms []db.Form) []*domain.Form {
	forms := make([]*domain.Form, len(dbForms))
	for i, dbForm := range dbForms {
		forms[i] = &domain.Form{}
		r.mapDBFormToDomain(dbForm, forms[i])
	}
	return forms
}
