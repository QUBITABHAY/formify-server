package response

import (
	"context"
	"log"
	"time"

	"formify/server/internal/integrations/google"
)

type FormGetter interface {
	GetFormByID(ctx context.Context, id int32) (schema []byte, sheetID *string, autoSync bool, userID int32, err error)
}

type UserTokenGetter interface {
	GetUserTokens(ctx context.Context, userID int32) (accessToken, refreshToken string, expiry time.Time, err error)
}

type Service struct {
	responseRepo    Repository
	sheetsService   *google.SheetsService
	formGetter      FormGetter
	userTokenGetter UserTokenGetter
}

func NewService(responseRepo Repository, sheetsService *google.SheetsService, formGetter FormGetter, userTokenGetter UserTokenGetter) *Service {
	return &Service{
		responseRepo:    responseRepo,
		sheetsService:   sheetsService,
		formGetter:      formGetter,
		userTokenGetter: userTokenGetter,
	}
}

func (s *Service) CreateResponse(ctx context.Context, response *Response) error {
	if err := s.responseRepo.Create(ctx, response); err != nil {
		return err
	}

	if s.formGetter != nil {
		go s.syncResponseToSheetIfEnabled(context.Background(), response)
	}

	return nil
}

func (s *Service) getSheetsServiceForUser(ctx context.Context, userID int32) *google.SheetsService {
	if s.userTokenGetter == nil {
		return s.sheetsService
	}

	accessToken, refreshToken, expiry, err := s.userTokenGetter.GetUserTokens(ctx, userID)
	if err != nil || accessToken == "" {
		log.Printf("Failed to get user tokens, falling back to service account: %v", err)
		return s.sheetsService
	}

	userSheetsService, err := google.NewSheetsServiceWithUserToken(ctx, accessToken, refreshToken, expiry)
	if err != nil {
		log.Printf("Failed to create user sheets service, falling back to service account: %v", err)
		return s.sheetsService
	}

	return userSheetsService
}

func (s *Service) syncResponseToSheetIfEnabled(ctx context.Context, response *Response) {
	schema, sheetID, autoSync, userID, err := s.formGetter.GetFormByID(ctx, response.FormID)
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

	sheetsService := s.getSheetsServiceForUser(ctx, userID)
	if sheetsService == nil {
		log.Printf("No sheets service available for sync")
		return
	}

	s.syncResponseToSheet(ctx, response, schema, *sheetID, sheetsService)
}

func (s *Service) syncResponseToSheet(ctx context.Context, response *Response, schema []byte, sheetID string, sheetsService *google.SheetsService) {
	fields, err := google.ParseFormSchema(schema)
	if err != nil {
		log.Printf("Failed to parse form schema for sheets sync: %v", err)
		row, _, err := google.ResponseToRowWithoutSchema(response.ID, response.CreatedAt, response.Data)
		if err != nil {
			log.Printf("Failed to convert response to row: %v", err)
			return
		}
		if err := sheetsService.AppendRow(ctx, sheetID, row); err != nil {
			log.Printf("Failed to append row to sheet: %v", err)
		}
		return
	}

	row, err := google.ResponseToRow(response.ID, response.CreatedAt, response.Data, fields)
	if err != nil {
		log.Printf("Failed to convert response to row: %v", err)
		return
	}

	if err := sheetsService.AppendRow(ctx, sheetID, row); err != nil {
		log.Printf("Failed to append row to sheet: %v", err)
	}
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
