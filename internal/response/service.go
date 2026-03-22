package response

import (
	"context"
	"fmt"
	"time"

	"formify/server/internal/integrations/google"
	"formify/server/internal/logger"

	"go.uber.org/zap"
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
		return nil
	}

	accessToken, refreshToken, expiry, err := s.userTokenGetter.GetUserTokens(ctx, userID)
	if err != nil || accessToken == "" {
		logger.GetLogger().Warn("Failed to get user OAuth tokens", zap.Error(err))
		return nil
	}

	userSheetsService, err := google.NewSheetsServiceWithUserToken(ctx, accessToken, refreshToken, expiry)
	if err != nil {
		logger.GetLogger().Warn("Failed to create user sheets service", zap.Error(err))
		return nil
	}

	return userSheetsService
}

func (s *Service) syncResponseToSheetIfEnabled(ctx context.Context, response *Response) {
	schema, sheetID, autoSync, userID, err := s.formGetter.GetFormByID(ctx, response.FormID)
	if err != nil {
		logger.GetLogger().Warn("Failed to get form for sheets sync", zap.Error(err))
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
		logger.GetLogger().Warn("No sheets service available for sync")
		return
	}

	s.syncResponseToSheet(ctx, response, schema, *sheetID, sheetsService)
}

func (s *Service) syncResponseToSheet(ctx context.Context, response *Response, schema []byte, sheetID string, sheetsService *google.SheetsService) {
	fields, err := google.ParseFormSchema(schema)
	if err != nil {
		logger.GetLogger().Warn("Failed to parse form schema for sheets sync", zap.Error(err))
		row, _, err := google.ResponseToRowWithoutSchema(response.ID, response.CreatedAt, response.Data)
		if err != nil {
			logger.GetLogger().Warn("Failed to convert response to row", zap.Error(err))
			return
		}
		if err := sheetsService.AppendRow(ctx, sheetID, row); err != nil {
			logger.GetLogger().Warn("Failed to append row to sheet", zap.Error(err))
		}
		return
	}

	row, err := google.ResponseToRow(response.ID, response.CreatedAt, response.Data, fields)
	if err != nil {
		logger.GetLogger().Warn("Failed to convert response to row", zap.Error(err))
		return
	}

	if err := sheetsService.AppendRow(ctx, sheetID, row); err != nil {
		logger.GetLogger().Warn("Failed to append row to sheet", zap.Error(err))
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
	response, err := s.responseRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.responseRepo.Delete(ctx, id); err != nil {
		return err
	}

	if s.formGetter == nil {
		return nil
	}

	go s.removeResponseFromSheetIfEnabled(context.Background(), response)

	return nil
}

func (s *Service) DeleteFormResponses(ctx context.Context, formID int32) error {
	return s.responseRepo.DeleteByFormID(ctx, formID)
}

func (s *Service) BackfillFormResponsesToSheet(ctx context.Context, formID int32, schema []byte, sheetID string, userID int32) error {
	if sheetID == "" {
		return nil
	}

	sheetsService := s.getSheetsServiceForUser(ctx, userID)
	if sheetsService == nil {
		return fmt.Errorf("no sheets service available")
	}

	responses, err := s.responseRepo.GetByFormID(ctx, formID)
	if err != nil {
		return fmt.Errorf("failed to load form responses: %w", err)
	}

	if len(responses) == 0 {
		return nil
	}

	existingIDs, err := sheetsService.GetExistingSubmissionIDs(ctx, sheetID)
	if err != nil {
		return fmt.Errorf("failed to read existing sheet rows: %w", err)
	}

	fields, schemaErr := google.ParseFormSchema(schema)

	for i := len(responses) - 1; i >= 0; i-- {
		resp := responses[i]
		if _, exists := existingIDs[resp.ID]; exists {
			continue
		}

		var row []interface{}
		if schemaErr == nil {
			row, err = google.ResponseToRow(resp.ID, resp.CreatedAt, resp.Data, fields)
		} else {
			row, _, err = google.ResponseToRowWithoutSchema(resp.ID, resp.CreatedAt, resp.Data)
		}
		if err != nil {
			logger.GetLogger().Warn("Failed to convert response to row for backfill", zap.Int32("response_id", resp.ID), zap.Error(err))
			continue
		}

		if err := sheetsService.AppendRow(ctx, sheetID, row); err != nil {
			logger.GetLogger().Warn("Failed to append response during backfill", zap.Int32("response_id", resp.ID), zap.Error(err))
			continue
		}
	}

	return nil
}

func (s *Service) removeResponseFromSheetIfEnabled(ctx context.Context, response *Response) {
	_, sheetID, autoSync, userID, err := s.formGetter.GetFormByID(ctx, response.FormID)
	if err != nil {
		logger.GetLogger().Warn("Failed to get form for sheets delete sync", zap.Error(err))
		return
	}

	if sheetID == nil || *sheetID == "" || !autoSync {
		return
	}

	sheetsService := s.getSheetsServiceForUser(ctx, userID)
	if sheetsService == nil {
		logger.GetLogger().Warn("No sheets service available for delete sync")
		return
	}

	if err := sheetsService.DeleteRowBySubmissionID(ctx, *sheetID, response.ID); err != nil {
		logger.GetLogger().Warn("Failed to remove response from sheet", zap.Int32("response_id", response.ID), zap.Error(err))
	}
}
