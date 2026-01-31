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

func (r *repository) GetByUserID(ctx context.Context, userID int32) ([]*Form, error) {
	dbForms, err := r.queries.ListFormsByUserID(ctx, userID)
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
