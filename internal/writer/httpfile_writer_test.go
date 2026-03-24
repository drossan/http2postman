package writer

import (
	"strings"
	"testing"

	"github.com/drossan/http2postman/internal/fs"
	"github.com/drossan/http2postman/internal/model"
)

func TestHTTPFileWriter_Write_SingleFile(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	w := NewHTTPFileWriter(memFS)

	files := []model.HTTPFile{
		{
			Path: "api/users.http",
			Requests: []model.HTTPRequest{
				{
					Name:    "Get Users",
					Method:  "GET",
					URL:     "https://api.example.com/users",
					Headers: []model.HTTPHeader{{Key: "Accept", Value: "application/json"}},
				},
			},
		},
	}

	err := w.Write(files, "output", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := memFS.ReadFile("output/api/users.http")
	if err != nil {
		t.Fatalf("file not written: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "# Get Users") {
		t.Error("expected request name in output")
	}
	if !strings.Contains(content, "GET https://api.example.com/users") {
		t.Error("expected method and URL in output")
	}
	if !strings.Contains(content, "Accept: application/json") {
		t.Error("expected header in output")
	}
}

func TestHTTPFileWriter_Write_MultipleFiles(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	w := NewHTTPFileWriter(memFS)

	files := []model.HTTPFile{
		{Path: "auth/login.http", Requests: []model.HTTPRequest{{Name: "Login", Method: "POST", URL: "http://x"}}},
		{Path: "users/list.http", Requests: []model.HTTPRequest{{Name: "List", Method: "GET", URL: "http://y"}}},
	}

	err := w.Write(files, "out", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !memFS.FileExists("out/auth/login.http") {
		t.Error("expected auth/login.http to exist")
	}
	if !memFS.FileExists("out/users/list.http") {
		t.Error("expected users/list.http to exist")
	}
}

func TestHTTPFileWriter_Write_OverwriteProtection(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	memFS.WriteFile("out/api.http", []byte("existing"), 0644)
	w := NewHTTPFileWriter(memFS)

	files := []model.HTTPFile{{Path: "api.http", Requests: []model.HTTPRequest{{Name: "T", Method: "GET", URL: "http://x"}}}}
	err := w.Write(files, "out", false)
	if err == nil {
		t.Fatal("expected error for existing file without force")
	}
}

func TestHTTPFileWriter_Write_ForceOverwrite(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	memFS.WriteFile("out/api.http", []byte("old"), 0644)
	w := NewHTTPFileWriter(memFS)

	files := []model.HTTPFile{{Path: "api.http", Requests: []model.HTTPRequest{{Name: "New", Method: "GET", URL: "http://x"}}}}
	err := w.Write(files, "out", true)
	if err != nil {
		t.Fatalf("unexpected error with force: %v", err)
	}

	data, _ := memFS.ReadFile("out/api.http")
	if string(data) == "old" {
		t.Error("file was not overwritten")
	}
}

func TestFormatHTTPRequest_GETNoBody(t *testing.T) {
	req := model.HTTPRequest{
		Name:    "Get Users",
		Method:  "GET",
		URL:     "https://api.example.com/users",
		Headers: []model.HTTPHeader{{Key: "Accept", Value: "application/json"}},
	}

	got := FormatHTTPRequest(req)
	expected := "# Get Users\nGET https://api.example.com/users\nAccept: application/json\n"
	if got != expected {
		t.Errorf("got:\n%s\nwant:\n%s", got, expected)
	}
}

func TestFormatHTTPRequest_POSTWithBody(t *testing.T) {
	req := model.HTTPRequest{
		Name:    "Create User",
		Method:  "POST",
		URL:     "https://api.example.com/users",
		Headers: []model.HTTPHeader{{Key: "Content-Type", Value: "application/json"}},
		Body:    `{"name":"John"}`,
	}

	got := FormatHTTPRequest(req)
	if !strings.Contains(got, "POST https://api.example.com/users") {
		t.Error("expected method and URL")
	}
	if !strings.Contains(got, `{"name":"John"}`) {
		t.Error("expected body")
	}
	if !strings.Contains(got, "Content-Type: application/json") {
		t.Error("expected header")
	}
}
