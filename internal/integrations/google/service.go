// Package google contains Google Sheets integration services and helpers.
package google

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"formify/server/internal/logger"
)

type SheetsService struct {
	service *sheets.Service
}

var (
	errMissingServiceAccountCreds = errors.New("google service account credentials are required")
	errNoAvailableSheets          = errors.New("spreadsheet has no available sheets")
)

func InitSheetsService(credentialsPath, credentialsJSON string) *SheetsService {
	if credentialsPath == "" && credentialsJSON == "" {
		return nil
	}

	sheetsService, err := NewSheetsService(context.Background(), credentialsPath, credentialsJSON)
	if err != nil {
		logger.GetLogger().Warn("Google Sheets integration not available", zap.Error(err))
		return nil
	}

	return sheetsService
}

func NewSheetsService(ctx context.Context, credentialsPath, credentialsJSON string) (*SheetsService, error) {
	scopes := []string{sheets.SpreadsheetsScope, drive.DriveScope}
	var opt option.ClientOption

	switch {
	case credentialsJSON != "":
		opt = option.WithAuthCredentialsJSON(option.ServiceAccount, []byte(credentialsJSON))
	case credentialsPath != "":
		opt = option.WithAuthCredentialsFile(option.ServiceAccount, credentialsPath)
	default:
		return nil, errMissingServiceAccountCreds
	}

	srv, err := sheets.NewService(ctx,
		opt,
		option.WithScopes(
			scopes...,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	return &SheetsService{service: srv}, nil
}

func NewSheetsServiceWithUserToken(ctx context.Context, accessToken, refreshToken string, expiry time.Time) (*SheetsService, error) {
	token := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       expiry,
	}

	tokenSource := oauth2.StaticTokenSource(token)

	srv, err := sheets.NewService(ctx,
		option.WithTokenSource(tokenSource),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service with user token: %w", err)
	}

	return &SheetsService{service: srv}, nil
}

func (s *SheetsService) CreateSpreadsheet(ctx context.Context, title string, headers []string) (string, error) {
	spreadsheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: title,
		},
		Sheets: []*sheets.Sheet{
			{
				Properties: &sheets.SheetProperties{
					Title: "Form Responses",
				},
			},
		},
	}

	created, err := s.service.Spreadsheets.Create(spreadsheet).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("failed to create spreadsheet: %w", err)
	}

	if len(headers) > 0 {
		headerRow := make([]any, len(headers))
		for i, h := range headers {
			headerRow[i] = h
		}

		valueRange := &sheets.ValueRange{
			Values: [][]any{headerRow},
		}

		_, err = s.service.Spreadsheets.Values.Update(
			created.SpreadsheetId,
			"Form Responses!A1",
			valueRange,
		).ValueInputOption("RAW").Context(ctx).Do()
		if err != nil {
			return "", fmt.Errorf("failed to add headers: %w", err)
		}
	}

	return created.SpreadsheetId, nil
}

func (s *SheetsService) AppendRow(ctx context.Context, spreadsheetID string, values []any) error {
	valueRange := &sheets.ValueRange{
		Values: [][]any{values},
	}

	_, err := s.service.Spreadsheets.Values.Append(
		spreadsheetID,
		"Form Responses!A:Z",
		valueRange,
	).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to append row: %w", err)
	}

	return nil
}

func (s *SheetsService) GetSpreadsheetTitle(ctx context.Context, spreadsheetID string) (string, error) {
	spreadsheet, err := s.service.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("failed to get spreadsheet: %w", err)
	}
	return spreadsheet.Properties.Title, nil
}

func (s *SheetsService) ValidateSpreadsheet(ctx context.Context, spreadsheetID string) error {
	_, err := s.service.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cannot access spreadsheet: %w", err)
	}
	return nil
}

func (s *SheetsService) GetExistingSubmissionIDs(ctx context.Context, spreadsheetID string) (map[int32]struct{}, error) {
	valueRange, err := s.service.Spreadsheets.Values.Get(spreadsheetID, "Form Responses!A:A").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read submission IDs: %w", err)
	}

	ids := make(map[int32]struct{})
	for i, row := range valueRange.Values {
		rowNumber := i + 1
		if rowNumber == 1 || len(row) == 0 {
			continue
		}

		raw := strings.TrimSpace(fmt.Sprintf("%v", row[0]))
		if raw == "" {
			continue
		}

		id64, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			continue
		}

		ids[int32(id64)] = struct{}{}
	}

	return ids, nil
}

func (s *SheetsService) DeleteRowBySubmissionID(ctx context.Context, spreadsheetID string, submissionID int32) error {
	valueRange, err := s.service.Spreadsheets.Values.Get(spreadsheetID, "Form Responses!A:A").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to read sheet rows: %w", err)
	}

	target := strconv.Itoa(int(submissionID))
	targetRow := findTargetRow(valueRange.Values, target)

	if targetRow == -1 {
		return nil
	}

	spreadsheet, err := s.service.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to load spreadsheet metadata: %w", err)
	}

	sheetID, err := resolveSheetID(spreadsheet)
	if err != nil {
		return err
	}

	_, err = s.service.Spreadsheets.BatchUpdate(spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				DeleteDimension: &sheets.DeleteDimensionRequest{
					Range: &sheets.DimensionRange{
						SheetId:    sheetID,
						Dimension:  "ROWS",
						StartIndex: int64(targetRow - 1),
						EndIndex:   int64(targetRow),
					},
				},
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to delete row for submission %d: %w", submissionID, err)
	}

	return nil
}

func findTargetRow(values [][]any, target string) int {
	for i, row := range values {
		rowNumber := i + 1
		if rowNumber == 1 || len(row) == 0 {
			continue
		}

		if strings.TrimSpace(fmt.Sprintf("%v", row[0])) == target {
			return rowNumber
		}
	}

	return -1
}

func resolveSheetID(spreadsheet *sheets.Spreadsheet) (int64, error) {
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties != nil && sheet.Properties.Title == "Form Responses" {
			return sheet.Properties.SheetId, nil
		}
	}

	if len(spreadsheet.Sheets) == 0 || spreadsheet.Sheets[0].Properties == nil {
		return 0, errNoAvailableSheets
	}

	return spreadsheet.Sheets[0].Properties.SheetId, nil
}
