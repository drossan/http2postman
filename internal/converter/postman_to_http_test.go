package converter

import (
	"testing"

	"github.com/drossan/http2postman/internal/model"
)

func TestCollectionToHTTPFiles_SimpleRequest(t *testing.T) {
	col := &model.PostmanCollection{
		Item: []model.PostmanItem{
			{
				Name: "Get Users",
				Request: &model.PostmanReq{
					Method: "GET",
					URL:    model.PostmanURL{Raw: "https://api.example.com/users"},
					Header: []model.PostmanHeader{{Key: "Accept", Value: "application/json"}},
				},
			},
		},
	}

	files := CollectionToHTTPFiles(col)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].Path != "get_users.http" {
		t.Errorf("path: got %q", files[0].Path)
	}
	if len(files[0].Requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(files[0].Requests))
	}
	req := files[0].Requests[0]
	if req.Method != "GET" {
		t.Errorf("method: got %q", req.Method)
	}
	if len(req.Headers) != 1 {
		t.Fatalf("expected 1 header, got %d", len(req.Headers))
	}
}

func TestCollectionToHTTPFiles_NestedFolders(t *testing.T) {
	col := &model.PostmanCollection{
		Item: []model.PostmanItem{
			{
				Name: "Auth",
				Item: []model.PostmanItem{
					{
						Name: "Login",
						Request: &model.PostmanReq{
							Method: "POST",
							URL:    model.PostmanURL{Raw: "https://api.example.com/login"},
						},
					},
				},
			},
		},
	}

	files := CollectionToHTTPFiles(col)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].Path != "auth/login.http" {
		t.Errorf("path: got %q, want %q", files[0].Path, "auth/login.http")
	}
}

func TestCollectionToHTTPFiles_AuthInheritance(t *testing.T) {
	col := &model.PostmanCollection{
		Item: []model.PostmanItem{
			{
				Name: "Protected",
				Auth: &model.PostmanAuth{
					Type:   "bearer",
					Bearer: []model.PostmanKV{{Key: "token", Value: "my-token"}},
				},
				Item: []model.PostmanItem{
					{
						Name: "Profile",
						Request: &model.PostmanReq{
							Method: "GET",
							URL:    model.PostmanURL{Raw: "https://api.example.com/profile"},
						},
					},
				},
			},
		},
	}

	files := CollectionToHTTPFiles(col)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	req := files[0].Requests[0]
	found := false
	for _, h := range req.Headers {
		if h.Key == "Authorization" && h.Value == "Bearer my-token" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected Authorization header with inherited token, got headers: %+v", req.Headers)
	}
}

func TestCollectionToHTTPFiles_RawBody(t *testing.T) {
	col := &model.PostmanCollection{
		Item: []model.PostmanItem{
			{
				Name: "Create",
				Request: &model.PostmanReq{
					Method: "POST",
					URL:    model.PostmanURL{Raw: "http://x"},
					Body:   &model.PostmanBody{Mode: "raw", Raw: `{"name":"test"}`},
				},
			},
		},
	}

	files := CollectionToHTTPFiles(col)
	if files[0].Requests[0].Body != `{"name":"test"}` {
		t.Errorf("body: got %q", files[0].Requests[0].Body)
	}
}

func TestCollectionToHTTPFiles_FormDataBody(t *testing.T) {
	col := &model.PostmanCollection{
		Item: []model.PostmanItem{
			{
				Name: "Upload",
				Request: &model.PostmanReq{
					Method: "POST",
					URL:    model.PostmanURL{Raw: "http://x"},
					Body: &model.PostmanBody{
						Mode: "formdata",
						FormData: []model.PostmanFormData{
							{Key: "name", Value: "John"},
							{Key: "email", Value: "john@test.com"},
							{Key: "", Value: "skip-empty"},
						},
					},
				},
			},
		},
	}

	files := CollectionToHTTPFiles(col)
	body := files[0].Requests[0].Body
	if body != "name: John\nemail: john@test.com" {
		t.Errorf("form body: got %q", body)
	}
}

func TestFormatFileName_TableDriven(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Get Users", "get_users"},
		{"Auth/Login", "auth_login"},
		{"Simple", "simple"},
		{"My API Endpoint", "my_api_endpoint"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := FormatFileName(tt.input)
			if got != tt.expected {
				t.Errorf("FormatFileName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExtractBearerToken_Valid(t *testing.T) {
	auth := &model.PostmanAuth{
		Type:   "bearer",
		Bearer: []model.PostmanKV{{Key: "token", Value: "abc123"}},
	}
	token, ok := ExtractBearerToken(auth)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if token != "abc123" {
		t.Errorf("token: got %q", token)
	}
}

func TestExtractBearerToken_NonBearer(t *testing.T) {
	auth := &model.PostmanAuth{Type: "basic"}
	_, ok := ExtractBearerToken(auth)
	if ok {
		t.Error("expected ok=false for non-bearer auth")
	}
}

func TestExtractBearerToken_Nil(t *testing.T) {
	_, ok := ExtractBearerToken(nil)
	if ok {
		t.Error("expected ok=false for nil auth")
	}
}
