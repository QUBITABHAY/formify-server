// Package fileupload provides file upload validation and storage integration.
package fileupload

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"

	"formify/server/internal/config"
)

const maxFileSizeBytes = 10 * 1024 * 1024
const fileBufSize = 512
const uploadIDLength = 16

var (
	errCloudinaryResponse       = errors.New("cloudinary response error")
	errCloudinaryDeleteResponse = errors.New("cloudinary delete returned unexpected result")
)

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

	buf := make([]byte, fileBufSize)
	n, err := file.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	contentType := http.DetectContentType(buf[:n])
	if _, seekErr := file.Seek(0, 0); seekErr != nil {
		return nil, fmt.Errorf("failed to seek file: %w", seekErr)
	}
	if !isAllowedMIMEType(contentType) {
		return nil, &ValidationError{Message: fmt.Sprintf("file type %q is not allowed", contentType)}
	}

	id, err := randomHex(uploadIDLength)
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
		return nil, fmt.Errorf("%w: %s", errCloudinaryResponse, resp.Error.Message)
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
		return fmt.Errorf("%w: %s", errCloudinaryDeleteResponse, resp.Result)
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

func isAllowedMIMEType(contentType string) bool {
	switch contentType {
	case "image/jpeg", "image/png", "image/gif", "image/webp", "application/pdf", "application/zip":
		return true
	default:
		return false
	}
}
