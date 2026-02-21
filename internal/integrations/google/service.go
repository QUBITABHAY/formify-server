package google

import (
	"context"
	"fmt"
	"log"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type SheetsService struct {
	service *sheets.Service
}

func InitSheetsService(credentialsPath string) *SheetsService {
	if credentialsPath == "" {
		return nil
	}

	sheetsService, err := NewSheetsService(context.Background(), credentialsPath)
	if err != nil {
		log.Printf("Warning: Google Sheets integration not available: %v", err)
		return nil
	}

	log.Println("Google Sheets integration enabled")
	return sheetsService
}

func NewSheetsService(ctx context.Context, credentialsPath string) (*SheetsService, error) {
	if credentialsPath == "" {
		return nil, fmt.Errorf("google service account key path is required")
	}

	srv, err := sheets.NewService(ctx,
		option.WithCredentialsFile(credentialsPath),
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
