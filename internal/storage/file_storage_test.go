package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileStorage_New_Success(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	if storage == nil {
		t.Fatalf("New() returned nil storage")
	}
	if storage.outputDir != tmpDir {
		t.Errorf("outputDir = %q, want %q", storage.outputDir, tmpDir)
	}
}

func TestFileStorage_New_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "new", "nested", "dir")
	storage, err := New(newDir)
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	if storage == nil {
		t.Fatalf("New() returned nil storage")
	}
	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		t.Errorf("Directory %q not created", newDir)
	}
}

func TestFileStorage_New_InvalidPath(t *testing.T) {
	invalidPath := "??"
	_, err := New(invalidPath)
	if err == nil {
		t.Errorf("New() expected error for invalid path, got nil")
	}
}

func TestFileStorage_Save_Success(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() failed %v", err)
	}

	content := "test file content"
	reader := strings.NewReader(content)

	path, err := storage.Save(reader, "test.txt")
	if err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "test.txt")
	if path != expectedPath {
		t.Errorf("path = %q, want %q", path, expectedPath)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() failed: %v", err)
	}
	if string(data) != content {
		t.Errorf("data = %q, want %q", data, content)
	}
}

func TestFileStorage_Save_EmptyContent(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() failed %v", err)
	}

	reader := strings.NewReader("")
	path, err := storage.Save(reader, "test.txt")

	if err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("File %q does not exist", path)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() failed: %v", err)
	}
	if info.Size() != 0 {
		t.Errorf("file size = %d, want 0", info.Size())
	}
}

func TestFileStorage_Save_LargeContent(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() failed %v", err)
	}
	content := strings.Repeat("a", 10*1024*1024)
	reader := strings.NewReader(content)

	path, err := storage.Save(reader, "test.txt")

	if err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() failed: %v", err)
	}
	if info.Size() != int64(len(content)) {
		t.Errorf("file size = %d, want %d", info.Size(), int64(len(content)))
	}
}

func TestFileStorage_Save_InvalidFilename(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() failed %v", err)
	}
	reader := strings.NewReader("test")
	_, err = storage.Save(reader, "???")
	if err == nil {
		t.Errorf("Save() expected error, got nil")
	}
}

func TestFileStorage_Save_DirectoryExists(t *testing.T) {
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "existing")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("MkdirAll() unexpected error: %v", err)
	}

	storage, err := New(subDir)
	if err != nil {
		t.Fatalf("New() failed %v", err)
	}

	content := "test file content"
	reader := strings.NewReader(content)

	path, err := storage.Save(reader, "test.txt")
	if err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() failed: %v", err)
	}
	if string(data) != content {
		t.Errorf("data = %q, want %q", string(data), content)
	}
}
