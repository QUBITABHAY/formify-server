package response

import (
	"context"
	"log"

	"formify/server/internal/integrations/google"
)

type FormGetter interface {
	GetFormByID(ctx context.Context, id int32) (schema []byte, sheetID *string, autoSync bool, err error)
}

type Service struct {
	responseRepo  Repository
	sheetsService *google.SheetsService
	formGetter    FormGetter
}

func NewService(responseRepo Repository, sheetsService *google.SheetsService, formGetter FormGetter) *Service {
	return &Service{
		responseRepo:  responseRepo,
		sheetsService: sheetsService,
		formGetter:    formGetter,
	}
}

func (s *Service) CreateResponse(ctx context.Context, response *Response) error {
	if err := s.responseRepo.Create(ctx, response); err != nil {
		return err
	}

	if s.sheetsService != nil && s.formGetter != nil {
		go s.syncResponseToSheetIfEnabled(context.Background(), response)
	}

	return nil
}

func (s *Service) syncResponseToSheetIfEnabled(ctx context.Context, response *Response) {
	schema, sheetID, autoSync, err := s.formGetter.GetFormByID(ctx, response.FormID)
	if err != nil {
		log.Printf("Failed to get form for sheets sync: %v", err)
		return
	}

	if sheetID == nil || *sheetID == "" {
		return
	}

	if !autoSync {
		return
	}

	s.syncResponseToSheet(ctx, response, schema, *sheetID)
}

func (s *Service) syncResponseToSheet(ctx context.Context, response *Response, schema []byte, sheetID string) {
	fields, err := google.ParseFormSchema(schema)
	if err != nil {
		log.Printf("Failed to parse form schema for sheets sync: %v", err)
		row, _, err := google.ResponseToRowWithoutSchema(response.ID, response.CreatedAt, response.Data)
		if err != nil {
			log.Printf("Failed to convert response to row: %v", err)
			return
		}
		if err := s.sheetsService.AppendRow(ctx, sheetID, row); err != nil {
			log.Printf("Failed to append row to sheet: %v", err)
		}
		return
	}

	row, err := google.ResponseToRow(response.ID, response.CreatedAt, response.Data, fields)
	if err != nil {
		log.Printf("Failed to convert response to row: %v", err)
		return
	}

	if err := s.sheetsService.AppendRow(ctx, sheetID, row); err != nil {
		log.Printf("Failed to append row to sheet: %v", err)
	}
}

func (s *Service) SyncResponseManually(ctx context.Context, responseID int32) error {
	if s.sheetsService == nil {
		return ErrSheetsNotConfigured
	}

	response, err := s.responseRepo.GetByID(ctx, responseID)
	if err != nil {
		return err
	}

	schema, sheetID, _, err := s.formGetter.GetFormByID(ctx, response.FormID)
	if err != nil {
		return err
	}

	if sheetID == nil || *sheetID == "" {
		return ErrNoSheetLinked
	}

	s.syncResponseToSheet(ctx, response, schema, *sheetID)
	return nil
}

func (s *Service) SyncAllResponses(ctx context.Context, formID int32) (int, error) {
	if s.sheetsService == nil {
		return 0, ErrSheetsNotConfigured
	}

	schema, sheetID, _, err := s.formGetter.GetFormByID(ctx, formID)
	if err != nil {
		return 0, err
	}

	if sheetID == nil || *sheetID == "" {
		return 0, ErrNoSheetLinked
	}

	responses, err := s.responseRepo.GetByFormID(ctx, formID)
	if err != nil {
		return 0, err
	}

	syncedCount := 0
	for _, response := range responses {
		s.syncResponseToSheet(ctx, response, schema, *sheetID)
		syncedCount++
	}

	return syncedCount, nil
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
