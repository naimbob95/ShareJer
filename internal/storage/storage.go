package storage

import "io"

type Storage interface {
	Save(filename string, src io.Reader) (storagePath string, err error)
	Open(storagePath string) (io.ReadCloser, error)
	Delete(storagePath string) error
	BasePath() string
}