package response

import (
	"context"
	"errors"
	"testing"
	"time"

	"formify/server/internal/integrations/google"
)

var (
	errNotFound    = errors.New("not found")
	errSchemaParse = errors.New("schema parse failed")
)

type mockRepository struct {
	deleteByFormIDFunc       func(ctx context.Context, formID int32) error
	getByIDFunc              func(ctx context.Context, id int32) (*Response, error)
	createFunc               func(ctx context.Context, response *Response) error
	deleteFunc               func(ctx context.Context, id int32) error
	getByFormIDPaginatedFunc func(ctx context.Context, formID, limit, offset int32) ([]*Response, error)
	getByFormIDFunc          func(ctx context.Context, formID int32) ([]*Response, error)
}

func (m *mockRepository) Create(ctx context.Context, response *Response) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, response)
	}
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id int32) (*Response, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockRepository) GetByFormID(ctx context.Context, formID int32) ([]*Response, error) {
	if m.getByFormIDFunc != nil {
		return m.getByFormIDFunc(ctx, formID)
	}
	return nil, nil
}

func (m *mockRepository) GetByFormIDPaginated(ctx context.Context, formID, limit, offset int32) ([]*Response, error) {
	if m.getByFormIDPaginatedFunc != nil {
		return m.getByFormIDPaginatedFunc(ctx, formID, limit, offset)
	}
	return nil, nil
}

func (m *mockRepository) Delete(ctx context.Context, id int32) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockRepository) DeleteByFormID(ctx context.Context, formID int32) error {
	if m.deleteByFormIDFunc != nil {
		return m.deleteByFormIDFunc(ctx, formID)
	}
	return nil
}

func TestService_CreateResponse(t *testing.T) {
	var savedResp *Response
	mockRepo := &mockRepository{
		createFunc: func(_ context.Context, resp *Response) error {
			savedResp = resp
			resp.ID = 100
			return nil
		},
	}

	svc := NewService(mockRepo, nil, nil, nil)
	r := &Response{FormID: 1, Data: []byte(`{"q1":"answer"}`)}
	err := svc.CreateResponse(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if savedResp.ID != 100 {
		t.Fatalf("expected ID 100, got %d", savedResp.ID)
	}
}

func TestService_GetResponseByID(t *testing.T) {
	mockRepo := &mockRepository{
		getByIDFunc: func(_ context.Context, id int32) (*Response, error) {
			if id == 100 {
				return &Response{ID: id, FormID: 1}, nil
			}
			return nil, errNotFound
		},
	}

	svc := NewService(mockRepo, nil, nil, nil)
	resp, err := svc.GetResponseByID(context.Background(), 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != 100 {
		t.Fatalf("expected ID 100, got %d", resp.ID)
	}
}

func TestService_GetFormResponses(t *testing.T) {
	mockRepo := &mockRepository{
		getByFormIDFunc: func(_ context.Context, formID int32) ([]*Response, error) {
			return []*Response{
				{ID: 1, FormID: formID},
				{ID: 2, FormID: formID},
			}, nil
		},
	}

	svc := NewService(mockRepo, nil, nil, nil)
	responses, err := svc.GetFormResponses(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(responses))
	}
}

func TestService_GetFormResponsesPaginated(t *testing.T) {
	mockRepo := &mockRepository{
		getByFormIDPaginatedFunc: func(_ context.Context, formID, _, _ int32) ([]*Response, error) {
			return []*Response{{ID: 1, FormID: formID}}, nil
		},
	}

	svc := NewService(mockRepo, nil, nil, nil)
	responses, err := svc.GetFormResponsesPaginated(context.Background(), 5, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
}

func TestService_DeleteResponse(t *testing.T) {
	var deletedID int32
	mockRepo := &mockRepository{
		getByIDFunc: func(_ context.Context, id int32) (*Response, error) {
			return &Response{ID: id, FormID: 1}, nil
		},
		deleteFunc: func(_ context.Context, id int32) error {
			deletedID = id
			return nil
		},
	}

	svc := NewService(mockRepo, nil, nil, nil)
	err := svc.DeleteResponse(context.Background(), 77)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedID != 77 {
		t.Fatalf("expected deletedID 77, got %d", deletedID)
	}
}

func TestService_DeleteFormResponses(t *testing.T) {
	var deletedFormID int32
	mockRepo := &mockRepository{
		deleteByFormIDFunc: func(_ context.Context, formID int32) error {
			deletedFormID = formID
			return nil
		},
	}

	svc := NewService(mockRepo, nil, nil, nil)
	err := svc.DeleteFormResponses(context.Background(), 12)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedFormID != 12 {
		t.Fatalf("expected deletedFormID 12, got %d", deletedFormID)
	}
}

func TestService_BackfillFormResponsesToSheet_EmptySheetID(t *testing.T) {
	svc := NewService(&mockRepository{}, nil, nil, nil)
	err := svc.BackfillFormResponsesToSheet(context.Background(), 1, []byte("[]"), "", 1)
	if err != nil {
		t.Fatalf("expected nil error for empty sheetID, got %v", err)
	}
}

func TestBuildBackfillRow(t *testing.T) {
	now := time.Now()
	resp := &Response{
		ID:        1,
		CreatedAt: now,
		Data:      []byte(`{"name":"Charlie"}`),
	}

	t.Run("with schema", func(t *testing.T) {
		fields := []google.FormField{
			{ID: "name", Label: "Name"},
		}
		row, err := buildBackfillRow(resp, fields, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(row) != 3 {
			t.Fatalf("expected 3 row items (ID, Date, Name), got %d", len(row))
		}
		if row[2] != "Charlie" {
			t.Fatalf("expected Charlie, got %v", row[2])
		}
	})

	t.Run("without schema on schema error", func(t *testing.T) {
		row, err := buildBackfillRow(resp, nil, errSchemaParse)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(row) != 3 {
			t.Fatalf("expected 3 row items, got %d", len(row))
		}
		if row[2] != "Charlie" {
			t.Fatalf("expected Charlie, got %v", row[2])
		}
	})
}
