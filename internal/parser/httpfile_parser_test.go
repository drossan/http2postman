package parser

import (
	"errors"
	"os"
	"testing"

	"github.com/drossan/http2postman/internal/fs"
	"github.com/drossan/http2postman/internal/model"
)

func TestParseHTTPContent_SingleGET(t *testing.T) {
	content, err := os.ReadFile("testdata/simple_get.http")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	requests, err := ParseHTTPContent(string(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(requests))
	}

	req := requests[0]
	if req.Name != "Get Users" {
		t.Errorf("name: got %q, want %q", req.Name, "Get Users")
	}
	if req.Method != "GET" {
		t.Errorf("method: got %q, want %q", req.Method, "GET")
	}
	if req.URL != "https://api.example.com/users" {
		t.Errorf("url: got %q, want %q", req.URL, "https://api.example.com/users")
	}
	if len(req.Headers) != 1 {
		t.Fatalf("expected 1 header, got %d", len(req.Headers))
	}
	if req.Headers[0].Key != "Accept" || req.Headers[0].Value != "application/json" {
		t.Errorf("header: got %+v", req.Headers[0])
	}
	if req.Body != "" {
		t.Errorf("expected empty body, got %q", req.Body)
	}
}

func TestParseHTTPContent_POSTWithHeadersAndBody(t *testing.T) {
	content, err := os.ReadFile("testdata/post_with_body.http")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	requests, err := ParseHTTPContent(string(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(requests))
	}

	req := requests[0]
	if req.Method != "POST" {
		t.Errorf("method: got %q, want %q", req.Method, "POST")
	}
	if len(req.Headers) != 2 {
		t.Fatalf("expected 2 headers, got %d", len(req.Headers))
	}
	if req.Body == "" {
		t.Fatal("expected non-empty body")
	}
	if req.Headers[0].Key != "Content-Type" {
		t.Errorf("first header key: got %q, want %q", req.Headers[0].Key, "Content-Type")
	}
}

func TestParseHTTPContent_MultipleRequests(t *testing.T) {
	content, err := os.ReadFile("testdata/multiple_requests.http")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	requests, err := ParseHTTPContent(string(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(requests) != 3 {
		t.Fatalf("expected 3 requests, got %d", len(requests))
	}

	methods := []string{"GET", "POST", "DELETE"}
	for i, m := range methods {
		if requests[i].Method != m {
			t.Errorf("request %d method: got %q, want %q", i, requests[i].Method, m)
		}
	}
}

func TestParseHTTPContent_EmptyContent(t *testing.T) {
	_, err := ParseHTTPContent("")
	if !errors.Is(err, model.ErrInvalidHTTPFormat) {
		t.Errorf("expected ErrInvalidHTTPFormat, got %v", err)
	}
}

func TestParseHTTPContent_Malformed(t *testing.T) {
	content, err := os.ReadFile("testdata/malformed.http")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	_, err = ParseHTTPContent(string(content))
	if !errors.Is(err, model.ErrInvalidURLFormat) {
		t.Errorf("expected ErrInvalidURLFormat, got %v", err)
	}
}

func TestHTTPFileParser_ParseFile(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	memFS.WriteFile("test/api.http", []byte("# Test\nGET https://example.com\n"), 0644)

	parser := NewHTTPFileParser(memFS)
	file, err := parser.ParseFile("test/api.http")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(file.Requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(file.Requests))
	}
	if file.Requests[0].Method != "GET" {
		t.Errorf("method: got %q, want %q", file.Requests[0].Method, "GET")
	}
}

func TestHTTPFileParser_ParseFile_NotFound(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	parser := NewHTTPFileParser(memFS)

	_, err := parser.ParseFile("nonexistent.http")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestHTTPFileParser_ParseDirectory(t *testing.T) {
	memFS := fs.NewMemoryFileSystem()
	memFS.WriteFile("root/api/users.http", []byte("# Users\nGET https://example.com/users\n"), 0644)
	memFS.WriteFile("root/api/auth.http", []byte("# Login\nPOST https://example.com/login\n"), 0644)
	memFS.WriteFile("root/readme.txt", []byte("not an http file"), 0644)

	parser := NewHTTPFileParser(memFS)
	files, err := parser.ParseDirectory("root")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 HTTP files, got %d", len(files))
	}
}
