package response

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"formify/server/internal/integrations/google"
	"formify/server/internal/logger"
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

var errNoSheetsService = errors.New("no sheets service available")

func NewService(
	responseRepo Repository,
	sheetsService *google.SheetsService,
	formGetter FormGetter,
	userTokenGetter UserTokenGetter,
) *Service {
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
		go s.syncResponseToSheetIfEnabled(ctx, response)
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
		row, _, rowErr := google.ResponseToRowWithoutSchema(response.ID, response.CreatedAt, response.Data)
		if rowErr != nil {
			logger.GetLogger().Warn("Failed to convert response to row", zap.Error(rowErr))
			return
		}
		if appendErr := sheetsService.AppendRow(ctx, sheetID, row); appendErr != nil {
			logger.GetLogger().Warn("Failed to append row to sheet", zap.Error(appendErr))
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

	go s.removeResponseFromSheetIfEnabled(ctx, response)

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
		return errNoSheetsService
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

	return s.appendMissingResponsesToSheet(ctx, sheetID, responses, existingIDs, fields, schemaErr, sheetsService)
}

func (s *Service) appendMissingResponsesToSheet(
	ctx context.Context,
	sheetID string,
	responses []*Response,
	existingIDs map[int32]struct{},
	fields []google.FormField,
	schemaErr error,
	sheetsService *google.SheetsService,
) error {
	for i := len(responses) - 1; i >= 0; i-- {
		resp := responses[i]
		if _, exists := existingIDs[resp.ID]; exists {
			continue
		}

		row, rowErr := buildBackfillRow(resp, fields, schemaErr)
		if rowErr != nil {
			logger.GetLogger().Warn("Failed to convert response to row for backfill", zap.Int32("response_id", resp.ID), zap.Error(rowErr))
			continue
		}

		if err := sheetsService.AppendRow(ctx, sheetID, row); err != nil {
			logger.GetLogger().Warn("Failed to append response during backfill", zap.Int32("response_id", resp.ID), zap.Error(err))
			continue
		}
	}

	return nil
}

func buildBackfillRow(resp *Response, fields []google.FormField, schemaErr error) ([]interface{}, error) {
	if schemaErr == nil {
		return google.ResponseToRow(resp.ID, resp.CreatedAt, resp.Data, fields)
	}

	row, _, err := google.ResponseToRowWithoutSchema(resp.ID, resp.CreatedAt, resp.Data)
	return row, err
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
