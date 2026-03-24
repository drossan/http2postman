package fs

import (
	"io/fs"
	"os"
	"path/filepath"
)

// OSFileSystem implements FileSystem using the real OS filesystem.
type OSFileSystem struct{}

func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

func (o *OSFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (o *OSFileSystem) WriteFile(path string, data []byte, perm fs.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	return os.WriteFile(path, data, perm)
}

func (o *OSFileSystem) Walk(root string, fn WalkFunc) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		return fn(path, info, err)
	})
}

func (o *OSFileSystem) Stat(path string) (fs.FileInfo, error) {
	return os.Stat(path)
}

func (o *OSFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (o *OSFileSystem) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
