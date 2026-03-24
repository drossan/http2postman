package parser

import (
	"errors"
	"os"
	"testing"

	"github.com/drossan/http2postman/internal/fs"
	"github.com/drossan/http2postman/internal/model"
)

func TestParseEnvironment_Valid(t *testing.T) {
	data, err := os.ReadFile("testdata/http-client.env.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	env, err := ParseEnvironment(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(env) != 2 {
		t.Fatalf("expected 2 environments, got %d", len(env))
	}
	if env["dev"]["host"] != "https://dev.api.example.com" {
		t.Errorf("dev host: got %q", env["dev"]["host"])
	}
	if env["prod"]["token"] != "prod-token-456" {
		t.Errorf("prod token: got %q", env["prod"]["token"])
	}
}

func TestParseEnvironment_InvalidJSON(t *testing.T) {
	_, err := ParseEnvironment([]byte(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestFindEnvFile_InCurrentDir(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	memFS.WriteFile("project/http-client.env.json", []byte(`{}`), 0644)

	path, err := FindEnvFile(memFS, "project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "project/http-client.env.json" {
		t.Errorf("path: got %q", path)
	}
}

func TestFindEnvFile_InParentDir(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	memFS.WriteFile("project/http-client.env.json", []byte(`{}`), 0644)
	memFS.MkdirAll("project/sub/deep", 0755)

	path, err := FindEnvFile(memFS, "project/sub/deep")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "project/http-client.env.json" {
		t.Errorf("path: got %q", path)
	}
}

func TestFindEnvFile_NotFound(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	memFS.MkdirAll("project", 0755)

	_, err := FindEnvFile(memFS, "project")
	if !errors.Is(err, model.ErrEnvFileNotFound) {
		t.Errorf("expected ErrEnvFileNotFound, got %v", err)
	}
}
