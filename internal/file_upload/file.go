package fileupload

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"mime/multipart"
	"net/http"

	"formify/server/internal/config"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

const maxFileSizeBytes = 10 * 1024 * 1024

var allowedMIMETypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"image/webp":      true,
	"application/pdf": true,
	"application/zip": true,
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

type Service struct {
	cld *cloudinary.Cloudinary
}

type UploadResult struct {
	PublicID string `json:"public_id"`
	URL      string `json:"url"`
	Format   string `json:"format"`
	Bytes    int    `json:"bytes"`
}

func NewService(cfg *config.Config) (*Service, error) {
	cld, err := cloudinary.NewFromParams(cfg.CloudinaryCloudName, cfg.CloudinaryAPIKey, cfg.CloudinaryAPISecret)
	if err != nil {
		return nil, fmt.Errorf("cloudinary init: %w", err)
	}
	cld.Config.URL.Secure = true
	return &Service{cld: cld}, nil
}

func (s *Service) UploadFile(ctx context.Context, formID string, file multipart.File, header *multipart.FileHeader) (*UploadResult, error) {
	if header.Size > maxFileSizeBytes {
		return nil, &ValidationError{Message: "file exceeds the maximum allowed size of 10 MB"}
	}

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	contentType := http.DetectContentType(buf[:n])
	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to seek file: %w", err)
	}
	if !allowedMIMETypes[contentType] {
		return nil, &ValidationError{Message: fmt.Sprintf("file type %q is not allowed", contentType)}
	}

	id, err := randomHex(16)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload ID: %w", err)
	}
	publicID := fmt.Sprintf("formify/%s/%s", formID, id)

	resp, err := s.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		PublicID:       publicID,
		UniqueFilename: api.Bool(false),
		Overwrite:      api.Bool(false),
		ResourceType:   "auto",
	})
	if err != nil {
		return nil, fmt.Errorf("cloudinary upload failed: %w", err)
	}
	if resp.Error.Message != "" {
		return nil, fmt.Errorf("cloudinary error: %s", resp.Error.Message)
	}

	return &UploadResult{
		PublicID: resp.PublicID,
		URL:      resp.SecureURL,
		Format:   resp.Format,
		Bytes:    resp.Bytes,
	}, nil
}

func (s *Service) DeleteFile(ctx context.Context, publicID string) error {
	resp, err := s.cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
	if err != nil {
		return fmt.Errorf("cloudinary delete failed: %w", err)
	}
	if resp.Result != "ok" {
		return fmt.Errorf("cloudinary delete returned unexpected result: %s", resp.Result)
	}
	return nil
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}