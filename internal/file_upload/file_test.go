package fileupload

import (
	"archive/zip"
	"bytes"
	"mime/multipart"
	"strings"
	"testing"
)

type memoryFile struct {
	*bytes.Reader
}

func (*memoryFile) Close() error { return nil }

func createTestZip(t *testing.T, files map[string][]byte) (file multipart.File, size int64) {
	t.Helper()
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("failed to create zip entry %s: %v", name, err)
		}
		if _, err := w.Write(content); err != nil {
			t.Fatalf("failed to write zip content: %v", err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}

	data := buf.Bytes()
	return &memoryFile{Reader: bytes.NewReader(data)}, int64(len(data))
}

func TestValidateZipArchive_CleanZip(t *testing.T) {
	file, size := createTestZip(t, map[string][]byte{
		"notes.txt":        []byte("hello world"),
		"images/chart.png": []byte("fake png bytes"),
	})

	err := validateZipArchive(file, size)
	if err != nil {
		t.Fatalf("expected clean zip to pass, got: %v", err)
	}
}

func TestValidateZipArchive_BlocksDangerousExtensions(t *testing.T) {
	dangerousFiles := []string{
		"payload.exe",
		"script.bat",
		"malware.sh",
		"exploit.vbs",
		"calc.cmd",
		"trojan.scr",
		"package.msi",
		"nested/sub/virus.js",
	}

	for _, badFile := range dangerousFiles {
		file, size := createTestZip(t, map[string][]byte{
			badFile: []byte("echo pwned"),
		})

		err := validateZipArchive(file, size)
		if err == nil {
			t.Errorf("expected file %s to be blocked, but it passed", badFile)
		} else if !strings.Contains(err.Error(), "restricted executable or script file") {
			t.Errorf("unexpected error for %s: %v", badFile, err)
		}
	}
}

func TestValidateZipArchive_BlocksDirectoryTraversal(t *testing.T) {
	file, size := createTestZip(t, map[string][]byte{
		"../../etc/passwd": []byte("root:x:0:0"),
	})

	err := validateZipArchive(file, size)
	if err == nil {
		t.Fatal("expected traversal zip to be rejected, but it passed")
	}
}
