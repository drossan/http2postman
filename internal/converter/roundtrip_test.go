package converter

import (
	"testing"

	"github.com/drossan/http2postman/internal/model"
	"github.com/drossan/http2postman/internal/parser"
	"github.com/drossan/http2postman/internal/writer"
)

func TestRoundtrip_ExportThenImport(t *testing.T) {
	// 1. Start with HTTP requests
	originalFiles := []model.HTTPFile{
		{
			Path: "backend/auth/login.http",
			Requests: []model.HTTPRequest{
				{
					Name:   "Login",
					Method: "POST",
					URL:    "https://api.example.com/login",
					Headers: []model.HTTPHeader{
						{Key: "Content-Type", Value: "application/json"},
					},
					Body: `{"user":"test","pass":"secret"}`,
				},
			},
		},
		{
			Path: "backend/users/list.http",
			Requests: []model.HTTPRequest{
				{
					Name:    "List Users",
					Method:  "GET",
					URL:     "https://api.example.com/users",
					Headers: []model.HTTPHeader{{Key: "Accept", Value: "application/json"}},
				},
			},
		},
	}

	// 2. Convert HTTP → Postman
	collection := HTTPFilesToCollection(originalFiles, "Roundtrip Test", "", nil)

	if collection.Info.Name != "Roundtrip Test" {
		t.Errorf("collection name: got %q", collection.Info.Name)
	}

	// 3. Convert Postman → HTTP
	importedFiles := CollectionToHTTPFiles(collection)

	if len(importedFiles) != 2 {
		t.Fatalf("expected 2 imported files, got %d", len(importedFiles))
	}

	// 4. Verify key data is preserved
	for _, imported := range importedFiles {
		if len(imported.Requests) != 1 {
			t.Errorf("file %s: expected 1 request, got %d", imported.Path, len(imported.Requests))
			continue
		}
		req := imported.Requests[0]

		// Find matching original
		var original *model.HTTPRequest
		for _, origFile := range originalFiles {
			for _, origReq := range origFile.Requests {
				if origReq.Name == req.Name {
					original = &origReq
					break
				}
			}
		}

		if original == nil {
			t.Errorf("no original found for imported request %q", req.Name)
			continue
		}

		if req.Method != original.Method {
			t.Errorf("request %q method: got %q, want %q", req.Name, req.Method, original.Method)
		}
		if req.URL != original.URL {
			t.Errorf("request %q URL: got %q, want %q", req.Name, req.URL, original.URL)
		}
		if req.Body != original.Body {
			t.Errorf("request %q body: got %q, want %q", req.Name, req.Body, original.Body)
		}
	}

	// 5. Verify formatting produces valid .http content
	for _, file := range importedFiles {
		content := writer.FormatHTTPFile(file)
		reparsed, err := parser.ParseHTTPContent(content)
		if err != nil {
			t.Errorf("re-parsing formatted file %s failed: %v", file.Path, err)
			continue
		}
		if len(reparsed) != len(file.Requests) {
			t.Errorf("file %s: re-parsed %d requests, expected %d", file.Path, len(reparsed), len(file.Requests))
		}
	}
}
