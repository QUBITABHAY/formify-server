package form

import (
	"context"
	"errors"
	"testing"

	"formify/server/internal/response"
)

var errResponseDeleteFailed = errors.New("response delete failed")

type mockFormRepository struct {
	createFunc            func(ctx context.Context, form *Form) error
	getByIDFunc           func(ctx context.Context, id int32) (*Form, error)
	getByShareURLFunc     func(ctx context.Context, shareURL string) (*Form, error)
	getByUserIDFunc       func(ctx context.Context, userID int32) ([]*Form, error)
	updateFunc            func(ctx context.Context, form *Form) error
	updateStatusFunc      func(ctx context.Context, id int32, status Status) (*Form, error)
	updateShareURLFunc    func(ctx context.Context, id int32, shareURL string) (*Form, error)
	deleteFunc            func(ctx context.Context, id int32) error
	linkGoogleSheetFunc   func(ctx context.Context, id int32, sheetID, sheetName string, autoSync bool) (*Form, error)
	unlinkGoogleSheetFunc func(ctx context.Context, id int32) (*Form, error)
}

func (m *mockFormRepository) Create(ctx context.Context, form *Form) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, form)
	}
	return nil
}

func (m *mockFormRepository) GetByID(ctx context.Context, id int32) (*Form, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockFormRepository) GetByShareURL(ctx context.Context, shareURL string) (*Form, error) {
	if m.getByShareURLFunc != nil {
		return m.getByShareURLFunc(ctx, shareURL)
	}
	return nil, nil
}

func (m *mockFormRepository) GetByUserID(ctx context.Context, userID int32) ([]*Form, error) {
	if m.getByUserIDFunc != nil {
		return m.getByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (*mockFormRepository) GetPublishedByUserID(_ context.Context, _ int32) ([]*Form, error) {
	return nil, nil
}

func (m *mockFormRepository) Update(ctx context.Context, form *Form) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, form)
	}
	return nil
}

func (m *mockFormRepository) UpdateStatus(ctx context.Context, id int32, status Status) (*Form, error) {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(ctx, id, status)
	}
	return nil, nil
}

func (m *mockFormRepository) UpdateShareURL(ctx context.Context, id int32, shareURL string) (*Form, error) {
	if m.updateShareURLFunc != nil {
		return m.updateShareURLFunc(ctx, id, shareURL)
	}
	return nil, nil
}

func (m *mockFormRepository) Delete(ctx context.Context, id int32) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockFormRepository) LinkGoogleSheet(ctx context.Context, id int32, sheetID, sheetName string, autoSync bool) (*Form, error) {
	if m.linkGoogleSheetFunc != nil {
		return m.linkGoogleSheetFunc(ctx, id, sheetID, sheetName, autoSync)
	}
	return nil, nil
}

func (m *mockFormRepository) UnlinkGoogleSheet(ctx context.Context, id int32) (*Form, error) {
	if m.unlinkGoogleSheetFunc != nil {
		return m.unlinkGoogleSheetFunc(ctx, id)
	}
	return nil, nil
}

type mockResponseRepository struct {
	deleteByFormIDFunc func(ctx context.Context, formID int32) error
}

func (*mockResponseRepository) Create(_ context.Context, _ *response.Response) error {
	return nil
}

func (*mockResponseRepository) GetByID(_ context.Context, _ int32) (*response.Response, error) {
	return nil, nil
}

func (*mockResponseRepository) GetByFormID(_ context.Context, _ int32) ([]*response.Response, error) {
	return nil, nil
}

func (*mockResponseRepository) GetByFormIDPaginated(_ context.Context, _, _, _ int32) ([]*response.Response, error) {
	return nil, nil
}

func (*mockResponseRepository) Delete(_ context.Context, _ int32) error {
	return nil
}

func (m *mockResponseRepository) DeleteByFormID(ctx context.Context, formID int32) error {
	if m.deleteByFormIDFunc != nil {
		return m.deleteByFormIDFunc(ctx, formID)
	}
	return nil
}

func TestService_CreateForm(t *testing.T) {
	var created *Form
	mockRepo := &mockFormRepository{
		createFunc: func(_ context.Context, f *Form) error {
			created = f
			return nil
		},
	}
	svc := NewService(mockRepo, nil)

	t.Run("applies defaults for status, schema, and settings", func(t *testing.T) {
		f := &Form{Name: "Contact Form"}
		err := svc.CreateForm(context.Background(), f)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created.Status != StatusDraft {
			t.Fatalf("expected StatusDraft, got %s", created.Status)
		}
		if string(created.Schema) != "[]" {
			t.Fatalf("expected default schema [], got %s", string(created.Schema))
		}
		if string(created.Settings) != "{}" {
			t.Fatalf("expected default settings {}, got %s", string(created.Settings))
		}
	})

	t.Run("preserves existing status, schema, and settings", func(t *testing.T) {
		f := &Form{
			Name:     "Existing Form",
			Status:   StatusPublished,
			Schema:   []byte(`[{"id":"1"}]`),
			Settings: []byte(`{"theme":"dark"}`),
		}
		err := svc.CreateForm(context.Background(), f)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created.Status != StatusPublished {
			t.Fatalf("expected StatusPublished, got %s", created.Status)
		}
		if string(created.Schema) != `[{"id":"1"}]` {
			t.Fatalf("expected preserved schema, got %s", string(created.Schema))
		}
	})
}

func TestService_GetFormByID(t *testing.T) {
	mockRepo := &mockFormRepository{
		getByIDFunc: func(_ context.Context, id int32) (*Form, error) {
			return &Form{ID: id, Name: "Test Form"}, nil
		},
	}
	svc := NewService(mockRepo, nil)

	f, err := svc.GetFormByID(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.ID != 10 {
		t.Fatalf("expected form id 10, got %d", f.ID)
	}
}

func TestService_GetFormByShareURL(t *testing.T) {
	mockRepo := &mockFormRepository{
		getByShareURLFunc: func(_ context.Context, shareURL string) (*Form, error) {
			return &Form{ID: 1, ShareURL: &shareURL}, nil
		},
	}
	svc := NewService(mockRepo, nil)

	f, err := svc.GetFormByShareURL(context.Background(), "my-share-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil || *f.ShareURL != "my-share-token" {
		t.Fatal("expected share URL to match")
	}
}

func TestService_PublishForm(t *testing.T) {
	t.Run("publishes and generates share URL when nil", func(t *testing.T) {
		var statusUpdated Status
		var shareURLUpdated string

		mockRepo := &mockFormRepository{
			updateStatusFunc: func(_ context.Context, id int32, status Status) (*Form, error) {
				statusUpdated = status
				return &Form{ID: id, Status: status, ShareURL: nil}, nil
			},
			updateShareURLFunc: func(_ context.Context, id int32, shareURL string) (*Form, error) {
				shareURLUpdated = shareURL
				return &Form{ID: id, Status: StatusPublished, ShareURL: &shareURL}, nil
			},
		}

		svc := NewService(mockRepo, nil)
		f, err := svc.PublishForm(context.Background(), 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if statusUpdated != StatusPublished {
			t.Fatalf("expected StatusPublished, got %s", statusUpdated)
		}
		if shareURLUpdated == "" {
			t.Fatal("expected share URL to be generated and updated")
		}
		if f.ShareURL == nil || *f.ShareURL != shareURLUpdated {
			t.Fatal("expected form to have updated share URL")
		}
	})

	t.Run("publishes without updating share URL when already present", func(t *testing.T) {
		existingURL := "already-exists"
		var shareURLUpdated bool

		mockRepo := &mockFormRepository{
			updateStatusFunc: func(_ context.Context, id int32, status Status) (*Form, error) {
				return &Form{ID: id, Status: status, ShareURL: &existingURL}, nil
			},
			updateShareURLFunc: func(_ context.Context, _ int32, _ string) (*Form, error) {
				shareURLUpdated = true
				return nil, nil
			},
		}

		svc := NewService(mockRepo, nil)
		f, err := svc.PublishForm(context.Background(), 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if shareURLUpdated {
			t.Fatal("should not update share URL if already present")
		}
		if *f.ShareURL != existingURL {
			t.Fatalf("expected existing share URL, got %s", *f.ShareURL)
		}
	})
}

func TestService_UnpublishForm(t *testing.T) {
	var statusUpdated Status
	mockRepo := &mockFormRepository{
		updateStatusFunc: func(_ context.Context, id int32, status Status) (*Form, error) {
			statusUpdated = status
			return &Form{ID: id, Status: status}, nil
		},
	}

	svc := NewService(mockRepo, nil)
	f, err := svc.UnpublishForm(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if statusUpdated != StatusDraft || f.Status != StatusDraft {
		t.Fatalf("expected StatusDraft, got %s", statusUpdated)
	}
}

func TestService_DeleteForm(t *testing.T) {
	t.Run("deletes responses and then form", func(t *testing.T) {
		var responsesDeleted bool
		var formDeleted bool

		mockResponseRepo := &mockResponseRepository{
			deleteByFormIDFunc: func(_ context.Context, _ int32) error {
				responsesDeleted = true
				return nil
			},
		}
		mockFormRepo := &mockFormRepository{
			deleteFunc: func(_ context.Context, _ int32) error {
				formDeleted = true
				return nil
			},
		}

		svc := NewService(mockFormRepo, mockResponseRepo)
		err := svc.DeleteForm(context.Background(), 5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !responsesDeleted || !formDeleted {
			t.Fatal("expected both responses and form to be deleted")
		}
	})

	t.Run("returns error if deleting responses fails", func(t *testing.T) {
		mockResponseRepo := &mockResponseRepository{
			deleteByFormIDFunc: func(_ context.Context, _ int32) error {
				return errResponseDeleteFailed
			},
		}
		mockFormRepo := &mockFormRepository{
			deleteFunc: func(_ context.Context, _ int32) error {
				return nil
			},
		}

		svc := NewService(mockFormRepo, mockResponseRepo)
		err := svc.DeleteForm(context.Background(), 5)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestService_GetFormOwnerID(t *testing.T) {
	mockRepo := &mockFormRepository{
		getByIDFunc: func(_ context.Context, id int32) (*Form, error) {
			return &Form{ID: id, UserID: 88}, nil
		},
	}
	svc := NewService(mockRepo, nil)
	ownerID, err := svc.GetFormOwnerID(context.Background(), 1)
	if err != nil || ownerID != 88 {
		t.Fatalf("expected ownerID 88, got %d (err: %v)", ownerID, err)
	}
}

func TestService_IsPublished(t *testing.T) {
	mockRepo := &mockFormRepository{
		getByIDFunc: func(_ context.Context, id int32) (*Form, error) {
			return &Form{ID: id, Status: StatusPublished}, nil
		},
	}
	svc := NewService(mockRepo, nil)
	pub, err := svc.IsPublished(context.Background(), 1)
	if err != nil || !pub {
		t.Fatalf("expected published true, got %v (err: %v)", pub, err)
	}
}

func TestService_GetFormSchema(t *testing.T) {
	mockRepo := &mockFormRepository{
		getByIDFunc: func(_ context.Context, id int32) (*Form, error) {
			return &Form{ID: id, Schema: []byte(`[{"type":"text"}]`)}, nil
		},
	}
	svc := NewService(mockRepo, nil)
	schema, err := svc.GetFormSchema(context.Background(), 1)
	if err != nil || string(schema) != `[{"type":"text"}]` {
		t.Fatalf("unexpected schema: %s", string(schema))
	}
}

func TestService_LinkAndUnlinkGoogleSheet(t *testing.T) {
	mockRepo := &mockFormRepository{
		linkGoogleSheetFunc: func(_ context.Context, id int32, sID, sName string, autoSync bool) (*Form, error) {
			return &Form{ID: id, GoogleSheetID: &sID, GoogleSheetName: &sName, GoogleSheetAutoSync: autoSync}, nil
		},
		unlinkGoogleSheetFunc: func(_ context.Context, id int32) (*Form, error) {
			return &Form{ID: id, GoogleSheetID: nil}, nil
		},
	}
	svc := NewService(mockRepo, nil)
	ctx := context.Background()

	linked, err := svc.LinkGoogleSheet(ctx, 1, "s1", "Responses", true)
	if err != nil || *linked.GoogleSheetID != "s1" {
		t.Fatalf("unexpected linked sheet: %+v", linked)
	}

	unlinked, err := svc.UnlinkGoogleSheet(ctx, 1)
	if err != nil || unlinked.GoogleSheetID != nil {
		t.Fatalf("expected unlinked sheet, got %+v", unlinked)
	}
}

func TestService_FormGetterAdapter(t *testing.T) {
	sheetID := "sheet-123"
	mockRepo := &mockFormRepository{
		getByIDFunc: func(_ context.Context, id int32) (*Form, error) {
			return &Form{
				ID:                  id,
				UserID:              88,
				Schema:              []byte(`[{"type":"text"}]`),
				GoogleSheetID:       &sheetID,
				GoogleSheetAutoSync: true,
			}, nil
		},
	}
	svc := NewService(mockRepo, nil)
	adapter := NewFormGetterAdapter(svc)

	adSchema, adSheetID, adAutoSync, adUserID, err := adapter.GetFormByID(context.Background(), 1)
	if err != nil || string(adSchema) != `[{"type":"text"}]` || *adSheetID != sheetID || !adAutoSync || adUserID != 88 {
		t.Fatalf("unexpected adapter values: %v, %v, %v, %v", string(adSchema), adSheetID, adAutoSync, adUserID)
	}
}
