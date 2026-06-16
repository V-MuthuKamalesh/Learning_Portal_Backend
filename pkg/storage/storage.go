// Package storage abstracts file uploads behind a driver interface so the
// application can switch between local disk and S3 without code changes.
package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/collegeassess/backend/configs"
	"github.com/google/uuid"
)

// Storage persists uploaded files and returns a public URL.
type Storage interface {
	Save(folder, filename string, r io.Reader) (publicURL string, err error)
}

// New returns the configured storage driver.
func New(cfg *configs.Config) Storage {
	switch cfg.Storage.Driver {
	case "s3":
		return &s3Storage{cfg: cfg.Storage}
	default:
		return &localStorage{dir: cfg.Storage.LocalDir, publicURL: cfg.Storage.PublicURL}
	}
}

// uniqueName prefixes a filename with a short UUID to avoid collisions.
func uniqueName(filename string) string {
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filepath.Base(filename), ext)
	base = strings.ReplaceAll(strings.ToLower(base), " ", "-")
	return fmt.Sprintf("%s-%s%s", base, uuid.NewString()[:8], ext)
}

// ── local disk ────────────────────────────────────────────────────────────────
type localStorage struct {
	dir       string
	publicURL string
}

func (s *localStorage) Save(folder, filename string, r io.Reader) (string, error) {
	name := uniqueName(filename)
	dir := filepath.Join(s.dir, folder)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	dst := filepath.Join(dir, name)
	f, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", strings.TrimRight(s.publicURL, "/"), folder, name), nil
}

// ── S3 ────────────────────────────────────────────────────────────────────────
// s3Storage is a placeholder wiring point. Add the aws-sdk-go-v2 dependency and
// implement Save with PutObject; the rest of the app already depends only on the
// Storage interface.
type s3Storage struct{ cfg configs.StorageConfig }

func (s *s3Storage) Save(folder, filename string, r io.Reader) (string, error) {
	return "", fmt.Errorf("s3 storage not yet implemented: add aws-sdk-go-v2 and PutObject in pkg/storage")
}
