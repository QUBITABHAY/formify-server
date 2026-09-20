// Package fileupload provides file upload validation and storage integration.
package fileupload

import (
	"archive/zip"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"

	"formify/server/internal/config"
)

const maxFileSizeBytes = 10 * 1024 * 1024
const fileBufSize = 512
const uploadIDLength = 16
const mimeZip = "application/zip"

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

func validateUploadFile(file multipart.File, header *multipart.FileHeader) error {
	if header.Size > maxFileSizeBytes {
		return &ValidationError{Message: "file exceeds the maximum allowed size of 10 MB"}
	}

	buf := make([]byte, fileBufSize)
	n, err := file.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	contentType := http.DetectContentType(buf[:n])
	if _, seekErr := file.Seek(0, 0); seekErr != nil {
		return fmt.Errorf("failed to seek file: %w", seekErr)
	}
	if !isAllowedMIMEType(contentType) {
		return &ValidationError{Message: fmt.Sprintf("file type %q is not allowed", contentType)}
	}

	if contentType == mimeZip {
		if zipErr := validateZipArchive(file, header.Size); zipErr != nil {
			return zipErr
		}
		if _, seekErr := file.Seek(0, 0); seekErr != nil {
			return fmt.Errorf("failed to reset file position: %w", seekErr)
		}
	}
	return nil
}

func (s *Service) UploadFile(ctx context.Context, formID string, file multipart.File, header *multipart.FileHeader) (*UploadResult, error) {
	if err := validateUploadFile(file, header); err != nil {
		return nil, err
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
	case "image/jpeg", "image/png", "image/gif", "image/webp", "application/pdf", mimeZip:
		return true
	default:
		return false
	}
}

func isDangerousZipExtension(ext string) bool {
	switch strings.ToLower(ext) {
	case ".exe", ".bat", ".cmd", ".sh",
		".vbs", ".vbe", ".js", ".jse",
		".wsf", ".wsh", ".scr", ".pif",
		".com", ".msi", ".dll", ".sys",
		".cpl", ".reg", ".ps1", ".ps2",
		".jar", ".apk", ".app", ".dmg",
		".iso", ".hta":
		return true
	default:
		return false
	}
}

func validateZipArchive(file multipart.File, size int64) error {
	zipReader, err := zip.NewReader(file, size)
	if err != nil {
		return &ValidationError{Message: "invalid or corrupted zip archive"}
	}

	const maxTotalUncompressedSize uint64 = 50 * 1024 * 1024 // 50 MB
	const maxFileCount = 500

	if len(zipReader.File) > maxFileCount {
		return &ValidationError{Message: fmt.Sprintf("zip contains too many files (maximum allowed is %d)", maxFileCount)}
	}

	var totalUncompressedSize uint64
	for _, f := range zipReader.File {
		// 1. Prevent Zip Slip / Directory Traversal in filenames
		cleanedPath := filepath.Clean(f.Name)
		if strings.HasPrefix(cleanedPath, "..") || filepath.IsAbs(cleanedPath) {
			return &ValidationError{Message: "zip contains invalid file paths"}
		}

		// Skip directories
		if f.FileInfo().IsDir() {
			continue
		}

		// 2. Block dangerous / executable file types
		ext := filepath.Ext(f.Name)
		if isDangerousZipExtension(ext) {
			return &ValidationError{
				Message: fmt.Sprintf("zip contains restricted executable or script file: %s", filepath.Base(f.Name)),
			}
		}

		// 3. Zip bomb check (uncompressed size sum)
		totalUncompressedSize += f.UncompressedSize64
		if totalUncompressedSize > maxTotalUncompressedSize {
			return &ValidationError{Message: "zip archive exceeds maximum uncompressed size of 50 MB"}
		}
	}

	return nil
}
