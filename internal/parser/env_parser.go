package parser

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/drossan/http2postman/internal/fs"
	"github.com/drossan/http2postman/internal/model"
)

// ParseEnvironment parses raw JSON bytes into an Environment.
func ParseEnvironment(data []byte) (model.Environment, error) {
	var env model.Environment
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("parsing environment JSON: %w", err)
	}
	return env, nil
}

// ParseEnvironmentFromFile reads and parses an environment JSON file.
func ParseEnvironmentFromFile(fsys fs.FileSystem, path string) (model.Environment, error) {
	data, err := fsys.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading environment file %s: %w", path, err)
	}
	return ParseEnvironment(data)
}

// FindEnvFile searches for http-client.env.json starting from startDir and walking up.
func FindEnvFile(fsys fs.FileSystem, startDir string) (string, error) {
	dir := filepath.Clean(startDir)
	for {
		envPath := filepath.Join(dir, "http-client.env.json")
		if fsys.FileExists(envPath) {
			return envPath, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", model.ErrEnvFileNotFound
}
