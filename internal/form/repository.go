package form

import (
	"context"
	"errors"

	"formify/server/internal/db"
	"formify/server/internal/shared"

	"github.com/jackc/pgx/v5"
)

var ErrFormNotFound = errors.New("form not found")

type repository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) Repository {
	return &repository{queries: queries}
}

func (r *repository) Create(ctx context.Context, form *Form) error {
	dbForm, err := r.queries.CreateForm(ctx, db.CreateFormParams{
		Name:        form.Name,
		Description: shared.StringToPgtypeText(form.Description),
		UserID:      form.UserID,
		Status:      db.NullFormStatus{FormStatus: db.FormStatus(form.Status), Valid: form.Status != ""},
		Schema:      form.Schema,
		Settings:    form.Settings,
		ShareUrl:    shared.StringToPgtypeText(form.ShareURL),
	})
	if err != nil {
		return err
	}
	r.mapDBFormToModel(dbForm, form)
	return nil
}

func (r *repository) GetByID(ctx context.Context, id int32) (*Form, error) {
	dbForm, err := r.queries.GetFormByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFormNotFound
		}
		return nil, err
	}
	form := &Form{}
	r.mapDBFormToModel(dbForm, form)
	return form, nil
}

func (r *repository) GetByShareURL(ctx context.Context, shareURL string) (*Form, error) {
	dbForm, err := r.queries.GetFormByShareURL(ctx, shared.StringToPgtypeText(&shareURL))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFormNotFound
		}
		return nil, err
	}
	form := &Form{}
	r.mapDBFormToModel(dbForm, form)
	return form, nil
}

func (r *repository) GetByUserID(ctx context.Context, userID int32) ([]*Form, error) {
	dbForms, err := r.queries.ListFormsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return r.mapDBFormsToModel(dbForms), nil
}

func (r *repository) GetPublishedByUserID(ctx context.Context, userID int32) ([]*Form, error) {
	dbForms, err := r.queries.ListPublishedFormsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return r.mapDBFormsToModel(dbForms), nil
}

func (r *repository) Update(ctx context.Context, form *Form) error {
	dbForm, err := r.queries.UpdateForm(ctx, db.UpdateFormParams{
		ID:          form.ID,
		Name:        form.Name,
		Description: shared.StringToPgtypeText(form.Description),
		Schema:      form.Schema,
		Settings:    form.Settings,
	})
	if err != nil {
		return err
	}
	r.mapDBFormToModel(dbForm, form)
	return nil
}

func (r *repository) UpdateStatus(ctx context.Context, id int32, status Status) (*Form, error) {
	dbForm, err := r.queries.UpdateFormStatus(ctx, db.UpdateFormStatusParams{
		ID:     id,
		Status: db.NullFormStatus{FormStatus: db.FormStatus(status), Valid: true},
	})
	if err != nil {
		return nil, err
	}
	form := &Form{}
	r.mapDBFormToModel(dbForm, form)
	return form, nil
}

func (r *repository) UpdateShareURL(ctx context.Context, id int32, shareURL string) (*Form, error) {
	dbForm, err := r.queries.UpdateFormShareURL(ctx, db.UpdateFormShareURLParams{
		ID:       id,
		ShareUrl: shared.StringToPgtypeText(&shareURL),
	})
	if err != nil {
		return nil, err
	}
	form := &Form{}
	r.mapDBFormToModel(dbForm, form)
	return form, nil
}

func (r *repository) Delete(ctx context.Context, id int32) error {
	return r.queries.DeleteForm(ctx, id)
}

func (r *repository) mapDBFormToModel(dbForm db.Form, form *Form) {
	form.ID = dbForm.ID
	form.FormID = shared.PgtypeTextToString(dbForm.FormID)
	form.Name = dbForm.Name
	form.Description = shared.PgtypeTextToString(dbForm.Description)
	form.UserID = dbForm.UserID
	if dbForm.Status.Valid {
		form.Status = Status(dbForm.Status.FormStatus)
	}
	form.Schema = dbForm.Schema
	form.Settings = dbForm.Settings
	form.ShareURL = shared.PgtypeTextToString(dbForm.ShareUrl)
	form.CreatedAt = shared.PgtypeTimestamptzToTime(dbForm.CreatedAt)
	form.UpdatedAt = shared.PgtypeTimestamptzToTime(dbForm.UpdatedAt)
}

func (r *repository) mapDBFormsToModel(dbForms []db.Form) []*Form {
	forms := make([]*Form, len(dbForms))
	for i, dbForm := range dbForms {
		forms[i] = &Form{}
		r.mapDBFormToModel(dbForm, forms[i])
	}
	return forms
}
