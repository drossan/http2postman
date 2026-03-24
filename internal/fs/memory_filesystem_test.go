package fs

import (
	"io/fs"
	"testing"
)

func TestMemoryFileSystem_ReadWriteFile(t *testing.T) {
	m := NewMemoryFileSystem()
	content := []byte("hello")

	if err := m.WriteFile("dir/test.txt", content, 0644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	got, err := m.ReadFile("dir/test.txt")
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("got %q, want %q", string(got), "hello")
	}
}

func TestMemoryFileSystem_ReadFile_NotFound(t *testing.T) {
	m := NewMemoryFileSystem()
	_, err := m.ReadFile("nonexistent.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestMemoryFileSystem_WriteFile_CreatesParentDirs(t *testing.T) {
	m := NewMemoryFileSystem()
	if err := m.WriteFile("a/b/c/file.txt", []byte("nested"), 0644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	if !m.Dirs["a"] {
		t.Error("expected parent dir 'a' to exist")
	}
	if !m.Dirs["a/b"] {
		t.Error("expected parent dir 'a/b' to exist")
	}
	if !m.Dirs["a/b/c"] {
		t.Error("expected parent dir 'a/b/c' to exist")
	}
}

func TestMemoryFileSystem_FileExists(t *testing.T) {
	m := NewMemoryFileSystem()

	if m.FileExists("file.txt") {
		t.Error("expected false before writing")
	}

	m.Files["file.txt"] = []byte("x")
	if !m.FileExists("file.txt") {
		t.Error("expected true after writing")
	}
}

func TestMemoryFileSystem_Walk_SortedOrder(t *testing.T) {
	m := NewMemoryFileSystem()
	m.WriteFile("root/b/file2.http", []byte("2"), 0644)
	m.WriteFile("root/a/file1.http", []byte("1"), 0644)

	var paths []string
	err := m.Walk("root", func(path string, info fs.FileInfo, err error) error {
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		t.Fatalf("Walk error: %v", err)
	}

	// Should be sorted: root, root/a, root/a/file1.http, root/b, root/b/file2.http
	for i := 1; i < len(paths); i++ {
		if paths[i] < paths[i-1] {
			t.Errorf("paths not sorted: %v", paths)
			break
		}
	}
}

func TestMemoryFileSystem_Stat_File(t *testing.T) {
	m := NewMemoryFileSystem()
	m.Files["file.txt"] = []byte("data")

	info, err := m.Stat("file.txt")
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

func TestMemoryFileSystem_Stat_Dir(t *testing.T) {
	m := NewMemoryFileSystem()
	m.MkdirAll("mydir", 0755)

	info, err := m.Stat("mydir")
	if err != nil {
		t.Fatalf("Stat error: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestMemoryFileSystem_Stat_NotFound(t *testing.T) {
	m := NewMemoryFileSystem()
	_, err := m.Stat("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestMemoryFileSystem_WriteFile_IsolatesCopy(t *testing.T) {
	m := NewMemoryFileSystem()
	original := []byte("original")
	m.WriteFile("file.txt", original, 0644)

	// Mutate the original slice
	original[0] = 'X'

	got, _ := m.ReadFile("file.txt")
	if string(got) != "original" {
		t.Errorf("WriteFile did not copy data: got %q", string(got))
	}
}
