package fs

import "io/fs"

// FileSystem abstracts filesystem operations for dependency injection and testing.
type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	Walk(root string, fn WalkFunc) error
	Stat(path string) (fs.FileInfo, error)
	MkdirAll(path string, perm fs.FileMode) error
	FileExists(path string) bool
}

// WalkFunc is the callback signature for Walk.
type WalkFunc func(path string, info fs.FileInfo, err error) error
