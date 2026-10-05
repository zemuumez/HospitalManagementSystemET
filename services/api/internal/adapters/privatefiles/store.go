package privatefiles

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"hms.local/api/internal/domain"
	"io"
	"os"
	"regexp"
)

type Store struct{ root *os.Root }

var keyPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

func New(directory string) (*Store, error) {
	if directory == "" {
		return nil, domain.ErrUnavailable
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return nil, err
	}
	return &Store{root: root}, nil
}
func (s *Store) Close() error { return s.root.Close() }
func (s *Store) Put(ctx context.Context, data []byte) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	key := hex.EncodeToString(random)
	f, err := s.root.OpenFile(key, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return "", err
	}
	_, err = f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		_ = s.root.Remove(key)
		return "", err
	}
	return key, nil
}
func (s *Store) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !keyPattern.MatchString(key) {
		return nil, domain.ErrNotFound
	}
	f, err := s.root.Open(key)
	if errors.Is(err, os.ErrNotExist) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		_ = f.Close()
		return nil, domain.ErrNotFound
	}
	return f, nil
}
func (s *Store) Remove(ctx context.Context, key string) error {
	if !keyPattern.MatchString(key) {
		return domain.ErrValidation
	}
	return s.root.Remove(key)
}
