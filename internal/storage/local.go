package storage

import (
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) (*LocalStorage, error) {
	if err := os.MkdirAll(basePath, 0750); err != nil {
		return nil, err
	}
	return &LocalStorage{basePath: basePath}, nil
}

func (s *LocalStorage) Save(filename string, src io.Reader) (string, error) {
	storagePath := filepath.Join(s.basePath, filename)

	dst, err := os.Create(storagePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(storagePath)
		return "", err
	}

	return storagePath, nil
}

func (s *LocalStorage) Open(storagePath string) (io.ReadCloser, error) {
	return os.Open(storagePath)
}

func (s *LocalStorage) Delete(storagePath string) error {
	return os.Remove(storagePath)
}

func (s *LocalStorage) BasePath() string {
	return s.basePath
}