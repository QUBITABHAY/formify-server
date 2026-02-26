package google

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type SheetsService struct {
	service *sheets.Service
}

func InitSheetsService(credentialsPath, credentialsJSON string) *SheetsService {
	if credentialsPath == "" && credentialsJSON == "" {
		return nil
	}

	sheetsService, err := NewSheetsService(context.Background(), credentialsPath, credentialsJSON)
	if err != nil {
		log.Printf("Warning: Google Sheets integration not available: %v", err)
		return nil
	}

	log.Println("Google Sheets integration enabled")
	return sheetsService
}

func NewSheetsService(ctx context.Context, credentialsPath, credentialsJSON string) (*SheetsService, error) {
	var opt option.ClientOption

	switch {
	case credentialsJSON != "":
		opt = option.WithCredentialsJSON([]byte(credentialsJSON))
	case credentialsPath != "":
		opt = option.WithCredentialsFile(credentialsPath)
	default:
		return nil, fmt.Errorf("google service account credentials are required")
	}

	srv, err := sheets.NewService(ctx,
		opt,
		option.WithScopes(
			sheets.SpreadsheetsScope,
			drive.DriveScope,
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
		headerRow := make([]interface{}, len(headers))
		for i, h := range headers {
			headerRow[i] = h
		}

		valueRange := &sheets.ValueRange{
			Values: [][]interface{}{headerRow},
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

func (s *SheetsService) AppendRow(ctx context.Context, spreadsheetID string, values []interface{}) error {
	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{values},
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
	targetRow := -1
	for i, row := range valueRange.Values {
		rowNumber := i + 1
		if rowNumber == 1 || len(row) == 0 {
			continue
		}

		if strings.TrimSpace(fmt.Sprintf("%v", row[0])) == target {
			targetRow = rowNumber
			break
		}
	}

	if targetRow == -1 {
		return nil
	}

	spreadsheet, err := s.service.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to load spreadsheet metadata: %w", err)
	}

	var sheetID int64
	found := false
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties != nil && sheet.Properties.Title == "Form Responses" {
			sheetID = sheet.Properties.SheetId
			found = true
			break
		}
	}

	if !found {
		if len(spreadsheet.Sheets) == 0 || spreadsheet.Sheets[0].Properties == nil {
			return fmt.Errorf("spreadsheet has no available sheets")
		}
		sheetID = spreadsheet.Sheets[0].Properties.SheetId
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
