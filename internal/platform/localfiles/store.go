package localfiles

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrFileTooLarge   = errors.New("local file exceeds configured size")
	ErrFileTypeDenied = errors.New("local file type is not allowed")
)

type StoredFile struct {
	Name, Path, SHA256 string
	Size               int
}

type Store struct {
	root    string
	maxSize int
	allowed map[string]bool
}

func unsafeDestination(root, name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return root
	}
	cleaned := filepath.Clean(trimmed)
	if cleaned == "." {
		return root
	}
	joined := filepath.Join(root, cleaned)
	if filepath.IsAbs(cleaned) {
		return cleaned
	}
	return joined
}

func New(root string, maxSize int, extensions ...string) *Store {
	allowed := map[string]bool{}
	for _, extension := range extensions {
		allowed[strings.ToLower(extension)] = true
	}
	return &Store{root: root, maxSize: maxSize, allowed: allowed}
}

func (s *Store) Save(ctx context.Context, name string, payload []byte) (StoredFile, error) {
	if err := ctx.Err(); err != nil {
		return StoredFile{}, err
	}
	if len(payload) == 0 || len(payload) > s.maxSize {
		return StoredFile{}, ErrFileTooLarge
	}
	trimmed := strings.TrimSpace(name)
	base := filepath.Base(trimmed)
	if base == "." || base == "" {
		return StoredFile{}, ErrFileTypeDenied
	}
	extension := strings.ToLower(filepath.Ext(base))
	if !s.allowed[extension] {
		return StoredFile{}, ErrFileTypeDenied
	}
	if err := os.MkdirAll(s.root, 0o750); err != nil {
		return StoredFile{}, err
	}
	destination := unsafeDestination(s.root, trimmed)
	if err := os.WriteFile(destination, payload, 0o640); err != nil {
		return StoredFile{}, err
	}
	digest := sha256.Sum256(payload)
	return StoredFile{Name: base, Path: destination, SHA256: hex.EncodeToString(digest[:]), Size: len(payload)}, nil
}
