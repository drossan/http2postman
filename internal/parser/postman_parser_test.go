package parser

import (
	"errors"
	"os"
	"testing"

	"github.com/drossan/http2postman/internal/fs"
	"github.com/drossan/http2postman/internal/model"
)

func TestParsePostmanCollection_SimpleRequest(t *testing.T) {
	data, err := os.ReadFile("testdata/simple_collection.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	col, err := ParsePostmanCollection(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.Info.Name != "Simple Collection" {
		t.Errorf("name: got %q, want %q", col.Info.Name, "Simple Collection")
	}
	if len(col.Item) != 1 {
		t.Fatalf("expected 1 item, got %d", len(col.Item))
	}
	if col.Item[0].Request.Method != "GET" {
		t.Errorf("method: got %q, want %q", col.Item[0].Request.Method, "GET")
	}
	if col.Item[0].Request.URL.Raw != "https://api.example.com/users" {
		t.Errorf("url: got %q", col.Item[0].Request.URL.Raw)
	}
}

func TestParsePostmanCollection_NestedFolders(t *testing.T) {
	data, err := os.ReadFile("testdata/nested_collection.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	col, err := ParsePostmanCollection(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(col.Item) != 2 {
		t.Fatalf("expected 2 top-level items, got %d", len(col.Item))
	}

	authFolder := col.Item[0]
	if !authFolder.IsFolder() {
		t.Error("expected Auth to be a folder")
	}
	if len(authFolder.Item) != 1 {
		t.Fatalf("expected 1 item in Auth folder, got %d", len(authFolder.Item))
	}
	if authFolder.Item[0].Request.Body == nil {
		t.Error("expected Login request to have a body")
	}
	if authFolder.Item[0].Request.Body.Mode != "raw" {
		t.Errorf("body mode: got %q, want %q", authFolder.Item[0].Request.Body.Mode, "raw")
	}
}

func TestParsePostmanCollection_BearerAuth(t *testing.T) {
	data, err := os.ReadFile("testdata/collection_with_auth.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	col, err := ParsePostmanCollection(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	folder := col.Item[0]
	if folder.Auth == nil {
		t.Fatal("expected folder to have auth")
	}
	if folder.Auth.Type != "bearer" {
		t.Errorf("auth type: got %q, want %q", folder.Auth.Type, "bearer")
	}
	if len(folder.Auth.Bearer) != 1 {
		t.Fatalf("expected 1 bearer entry, got %d", len(folder.Auth.Bearer))
	}
	if folder.Auth.Bearer[0].Value != "my-secret-token" {
		t.Errorf("token: got %q", folder.Auth.Bearer[0].Value)
	}
}

func TestParsePostmanCollection_URLAsString(t *testing.T) {
	data, err := os.ReadFile("testdata/simple_collection.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	col, err := ParsePostmanCollection(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.Item[0].Request.URL.Raw != "https://api.example.com/users" {
		t.Errorf("URL from string: got %q", col.Item[0].Request.URL.Raw)
	}
}

func TestParsePostmanCollection_URLAsObject(t *testing.T) {
	data, err := os.ReadFile("testdata/nested_collection.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	col, err := ParsePostmanCollection(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	loginReq := col.Item[0].Item[0]
	if loginReq.Request.URL.Raw != "https://api.example.com/login" {
		t.Errorf("URL from object: got %q", loginReq.Request.URL.Raw)
	}
}

func TestParsePostmanCollection_InvalidJSON(t *testing.T) {
	_, err := ParsePostmanCollection([]byte(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParsePostmanCollection_MissingItems(t *testing.T) {
	data := []byte(`{"info":{"name":"Empty"}}}`)
	_, err := ParsePostmanCollection(data)
	if err == nil {
		t.Fatal("expected error for missing items")
	}
}

func TestParsePostmanCollectionFromFile(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	jsonData := `{"info":{"name":"Test","schema":"v2.1"},"item":[{"name":"Req","request":{"method":"GET","url":"http://x"}}]}`
	memFS.WriteFile("col.json", []byte(jsonData), 0644)

	col, err := ParsePostmanCollectionFromFile(memFS, "col.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.Info.Name != "Test" {
		t.Errorf("name: got %q", col.Info.Name)
	}
}

func TestParsePostmanCollectionFromFile_NotFound(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	_, err := ParsePostmanCollectionFromFile(memFS, "nonexistent.json")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParsePostmanCollection_EmptyItemField(t *testing.T) {
	data := []byte(`{"info":{"name":"NoItems","schema":"v2.1"}}`)
	_, err := ParsePostmanCollection(data)
	if !errors.Is(err, model.ErrInvalidCollection) {
		t.Errorf("expected ErrInvalidCollection, got %v", err)
	}
}
