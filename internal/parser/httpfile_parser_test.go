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
	if !errors.Is(err, model.ErrInvalidHTTPFormat) {
		t.Errorf("expected ErrInvalidHTTPFormat, got %v", err)
	}
}

func TestParseHTTPContent_NoCommentWithHost(t *testing.T) {
	content, err := os.ReadFile("testdata/no_comment_with_host.http")
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
	// URL should be Host + path
	if req.URL != "{{API_URL}}/cache/public-api/invalidate" {
		t.Errorf("url: got %q, want %q", req.URL, "{{API_URL}}/cache/public-api/invalidate")
	}
	// Host header should be filtered out
	for _, h := range req.Headers {
		if h.Key == "Host" {
			t.Error("Host header should be filtered out")
		}
	}
	if len(req.Headers) != 2 {
		t.Errorf("expected 2 headers (Content-Type, Authorization), got %d: %+v", len(req.Headers), req.Headers)
	}
	if req.Body == "" {
		t.Fatal("expected non-empty body")
	}
	// Name is auto-generated from method + path (before Host resolution)
	if req.Name != "POST /cache/public-api/invalidate" {
		t.Errorf("name: got %q", req.Name)
	}
}

func TestParseHTTPContent_NoCommentAbsoluteURL(t *testing.T) {
	content, err := os.ReadFile("testdata/no_comment_absolute_url.http")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	requests, err := ParseHTTPContent(string(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := requests[0]
	if req.Method != "GET" {
		t.Errorf("method: got %q, want %q", req.Method, "GET")
	}
	// HTTP/1.1 suffix should be stripped, URL preserved
	if req.URL != "https://api.example.com/users" {
		t.Errorf("url: got %q, want %q", req.URL, "https://api.example.com/users")
	}
	if len(req.Headers) != 2 {
		t.Errorf("expected 2 headers, got %d", len(req.Headers))
	}
}

func TestParseHTTPContent_MixedFormats(t *testing.T) {
	content, err := os.ReadFile("testdata/mixed_formats.http")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	requests, err := ParseHTTPContent(string(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(requests))
	}

	// First: with comment
	if requests[0].Name != "Get Users" {
		t.Errorf("first request name: got %q", requests[0].Name)
	}
	if requests[0].Method != "GET" {
		t.Errorf("first request method: got %q", requests[0].Method)
	}

	// Second: without comment, with Host
	if requests[1].Method != "POST" {
		t.Errorf("second request method: got %q", requests[1].Method)
	}
	if requests[1].URL != "{{API_URL}}/auth/login" {
		t.Errorf("second request url: got %q", requests[1].URL)
	}
}

func TestParseHTTPContent_MultilineComments(t *testing.T) {
	content, err := os.ReadFile("testdata/multiline_comments.http")
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
	// First comment line becomes the name
	if req.Name != "Returns all available activity event type categories for filtering." {
		t.Errorf("name: got %q", req.Name)
	}
	if req.Method != "GET" {
		t.Errorf("method: got %q", req.Method)
	}
	if req.URL != "{{API_URL}}/logs/activity/events-type" {
		t.Errorf("url: got %q", req.URL)
	}
	// Host should be filtered out, leaving Content-Type and Authorization
	if len(req.Headers) != 2 {
		t.Errorf("expected 2 headers, got %d: %+v", len(req.Headers), req.Headers)
	}
}

func TestCleanCommentName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"# Create Social Token", "Create Social Token"},
		{"# ─── Create Social Token ────────────────────────────────────", "Create Social Token"},
		{"# --- My Request ---", "My Request"},
		{"# === My Request ===", "My Request"},
		{"# ─── Get Users ───", "Get Users"},
		{"# Simple Name", "Simple Name"},
		{"## Double Hash", "Double Hash"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := cleanCommentName(tt.input)
			if got != tt.expected {
				t.Errorf("cleanCommentName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseHTTPContent_DecorativeComments(t *testing.T) {
	content := "# ─── Create Social Token ────────────────────────────────────\nPOST https://api.example.com/token\nContent-Type: application/json\n"

	requests, err := ParseHTTPContent(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requests[0].Name != "Create Social Token" {
		t.Errorf("name: got %q, want %q", requests[0].Name, "Create Social Token")
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
