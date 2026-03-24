package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOSFileSystem_ReadWriteFile(t *testing.T) {
	dir := t.TempDir()
	fsys := NewOSFileSystem()
	path := filepath.Join(dir, "test.txt")
	content := []byte("hello world")

	if err := fsys.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	got, err := fsys.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("got %q, want %q", string(got), string(content))
	}
}

func TestOSFileSystem_ReadFile_NotFound(t *testing.T) {
	fsys := NewOSFileSystem()
	_, err := fsys.ReadFile("/nonexistent/path/file.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestOSFileSystem_WriteFile_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	fsys := NewOSFileSystem()
	path := filepath.Join(dir, "a", "b", "c", "file.txt")

	if err := fsys.WriteFile(path, []byte("nested"), 0644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile error: %v", err)
	}
	if string(got) != "nested" {
		t.Errorf("got %q, want %q", string(got), "nested")
	}
}

func TestOSFileSystem_FileExists(t *testing.T) {
	dir := t.TempDir()
	fsys := NewOSFileSystem()
	path := filepath.Join(dir, "exists.txt")

	if fsys.FileExists(path) {
		t.Error("expected FileExists to return false before file creation")
	}

	if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
		t.Fatalf("os.WriteFile error: %v", err)
	}

	if !fsys.FileExists(path) {
		t.Error("expected FileExists to return true after file creation")
	}
}

func TestOSFileSystem_Walk(t *testing.T) {
	dir := t.TempDir()
	fsys := NewOSFileSystem()

	// Create structure: dir/a/file1.txt, dir/b/file2.txt
	if err := os.MkdirAll(filepath.Join(dir, "a"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "b"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a", "file1.txt"), []byte("1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b", "file2.txt"), []byte("2"), 0644); err != nil {
		t.Fatal(err)
	}

	var files []string
	err := fsys.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, _ := filepath.Rel(dir, path)
			files = append(files, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Walk error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(files), files)
	}
}

func TestOSFileSystem_Stat(t *testing.T) {
	dir := t.TempDir()
	fsys := NewOSFileSystem()
	path := filepath.Join(dir, "stat.txt")

	if err := os.WriteFile(path, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := fsys.Stat(path)
	if err != nil {
		t.Fatalf("Stat error: %v", err)
	}
	if info.IsDir() {
		t.Error("expected file, not directory")
	}
	if info.Size() != 4 {
		t.Errorf("expected size 4, got %d", info.Size())
	}
}
