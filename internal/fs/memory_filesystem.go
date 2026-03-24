package fs

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// MemoryFileSystem implements FileSystem in memory for testing.
type MemoryFileSystem struct {
	Files map[string][]byte
	Dirs  map[string]bool
}

func NewMemoryFileSystem() *MemoryFileSystem {
	return &MemoryFileSystem{
		Files: make(map[string][]byte),
		Dirs:  make(map[string]bool),
	}
}

func (m *MemoryFileSystem) ReadFile(path string) ([]byte, error) {
	data, ok := m.Files[path]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	return data, nil
}

func (m *MemoryFileSystem) WriteFile(path string, data []byte, _ fs.FileMode) error {
	dir := filepath.Dir(path)
	m.mkdirAllInternal(dir)
	m.Files[path] = make([]byte, len(data))
	copy(m.Files[path], data)
	return nil
}

func (m *MemoryFileSystem) Walk(root string, fn WalkFunc) error {
	// Collect all paths under root
	var paths []string
	cleanRoot := filepath.Clean(root)

	for dir := range m.Dirs {
		if dir == cleanRoot || strings.HasPrefix(dir, cleanRoot+string(filepath.Separator)) {
			paths = append(paths, dir)
		}
	}
	for file := range m.Files {
		if file == cleanRoot || strings.HasPrefix(file, cleanRoot+string(filepath.Separator)) {
			paths = append(paths, file)
		}
	}

	sort.Strings(paths)

	for _, p := range paths {
		var info fs.FileInfo
		if _, isFile := m.Files[p]; isFile {
			info = &memFileInfo{name: filepath.Base(p), size: int64(len(m.Files[p])), isDir: false}
		} else {
			info = &memFileInfo{name: filepath.Base(p), isDir: true}
		}
		if err := fn(p, info, nil); err != nil {
			return err
		}
	}
	return nil
}

func (m *MemoryFileSystem) Stat(path string) (fs.FileInfo, error) {
	if data, ok := m.Files[path]; ok {
		return &memFileInfo{name: filepath.Base(path), size: int64(len(data)), isDir: false}, nil
	}
	if m.Dirs[path] {
		return &memFileInfo{name: filepath.Base(path), isDir: true}, nil
	}
	return nil, fmt.Errorf("not found: %s", path)
}

func (m *MemoryFileSystem) MkdirAll(path string, _ fs.FileMode) error {
	m.mkdirAllInternal(path)
	return nil
}

func (m *MemoryFileSystem) FileExists(path string) bool {
	_, hasFile := m.Files[path]
	return hasFile || m.Dirs[path]
}

func (m *MemoryFileSystem) mkdirAllInternal(path string) {
	parts := strings.Split(filepath.Clean(path), string(filepath.Separator))
	for i := range parts {
		dir := strings.Join(parts[:i+1], string(filepath.Separator))
		if dir != "" {
			m.Dirs[dir] = true
		}
	}
}

// memFileInfo implements fs.FileInfo for in-memory files.
type memFileInfo struct {
	name  string
	size  int64
	isDir bool
}

func (f *memFileInfo) Name() string      { return f.name }
func (f *memFileInfo) Size() int64       { return f.size }
func (f *memFileInfo) Mode() fs.FileMode { return 0644 }
func (f *memFileInfo) ModTime() time.Time { return time.Time{} }
func (f *memFileInfo) IsDir() bool       { return f.isDir }
func (f *memFileInfo) Sys() interface{}  { return nil }
