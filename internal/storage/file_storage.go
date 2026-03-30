package storage

import (
	"fmt"
	"io"
	"os"
	fp "path/filepath"
)

type FileStorage struct {
	outputDir string
}

func New(outputDir string) (*FileStorage, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating output directory %s: %s", outputDir, err)
	}
	return &FileStorage{outputDir: outputDir}, nil
}

func (s *FileStorage) Save(r io.Reader, filename string) (string, error) {
	filepath := fp.Join(s.outputDir, filename)

	f, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("error creating file %s: %s", filepath, err)
	}

	_, err = io.Copy(f, r)
	if err != nil {
		// Закрываем файл при ошибке копирования
		closeErr := f.Close()
		if closeErr != nil {
			return "", fmt.Errorf("error saving file %s: %w; also failed to close: %v", filepath, err, closeErr)
		}
		return "", fmt.Errorf("error saving file %s: %w", filepath, err)
	}

	if err := f.Close(); err != nil {
		return "", fmt.Errorf("error closing file %s: %w", filepath, err)
	}
	return filepath, nil
}
