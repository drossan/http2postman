package converter

import (
	"testing"

	"github.com/drossan/http2postman/internal/model"
)

func TestHTTPFilesToCollection_SingleFile(t *testing.T) {
	files := []model.HTTPFile{
		{
			Path: "api.http",
			Requests: []model.HTTPRequest{
				{Name: "Get Users", Method: "GET", URL: "https://api.example.com/users"},
			},
		},
	}

	col := HTTPFilesToCollection(files, "Test", nil)
	if col.Info.Name != "Test" {
		t.Errorf("name: got %q", col.Info.Name)
	}
	if col.Info.Schema != model.PostmanSchemaV210 {
		t.Errorf("schema: got %q", col.Info.Schema)
	}
	if len(col.Item) != 1 {
		t.Fatalf("expected 1 item, got %d", len(col.Item))
	}
	if len(col.Item[0].Item) != 1 {
		t.Fatalf("expected 1 request in group, got %d", len(col.Item[0].Item))
	}
	if col.Item[0].Item[0].Request.Method != "GET" {
		t.Errorf("method: got %q", col.Item[0].Item[0].Request.Method)
	}
}

func TestHTTPFilesToCollection_DirectoryHierarchy(t *testing.T) {
	files := []model.HTTPFile{
		{
			Path: "backend/auth/login.http",
			Requests: []model.HTTPRequest{
				{Name: "Login", Method: "POST", URL: "https://api.example.com/login"},
			},
		},
		{
			Path: "backend/users/list.http",
			Requests: []model.HTTPRequest{
				{Name: "List Users", Method: "GET", URL: "https://api.example.com/users"},
			},
		},
	}

	col := HTTPFilesToCollection(files, "Test", nil)

	// Should have one top-level folder: "Backend"
	if len(col.Item) != 1 {
		t.Fatalf("expected 1 top-level item, got %d", len(col.Item))
	}
	if col.Item[0].Name != "Backend" {
		t.Errorf("top folder: got %q, want %q", col.Item[0].Name, "Backend")
	}

	// Backend should have 2 subfolders: Auth and Users
	if len(col.Item[0].Item) != 2 {
		t.Fatalf("expected 2 subfolders, got %d", len(col.Item[0].Item))
	}
}

func TestHTTPFilesToCollection_WithEnvironment(t *testing.T) {
	files := []model.HTTPFile{
		{
			Path:     "api.http",
			Requests: []model.HTTPRequest{{Name: "Test", Method: "GET", URL: "http://x"}},
		},
	}
	env := &model.Environment{
		"dev": {"host": "https://dev.api.com", "token": "abc"},
	}

	col := HTTPFilesToCollection(files, "Test", env)
	if len(col.Variable) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(col.Variable))
	}
}

func TestFormatGroupName_TableDriven(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"auth_users", "Auth Users"},
		{"my-api", "My Api"},
		{"simple", "Simple"},
		{"auth_users.http", "Auth Users"},
		{"hello_world-test", "Hello World Test"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := FormatGroupName(tt.input)
			if got != tt.expected {
				t.Errorf("FormatGroupName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestHTTPRequestToPostmanItem_WithBody(t *testing.T) {
	req := model.HTTPRequest{
		Name:   "Create User",
		Method: "POST",
		URL:    "https://api.example.com/users",
		Body:   `{"name":"John"}`,
	}

	item := httpRequestToPostmanItem(req)
	if item.Request.Body == nil {
		t.Fatal("expected body")
	}
	if item.Request.Body.Mode != "raw" {
		t.Errorf("body mode: got %q", item.Request.Body.Mode)
	}
	if item.Request.Body.Raw != `{"name":"John"}` {
		t.Errorf("body raw: got %q", item.Request.Body.Raw)
	}
}

func TestHTTPRequestToPostmanItem_WithHeaders(t *testing.T) {
	req := model.HTTPRequest{
		Name:   "Test",
		Method: "GET",
		URL:    "http://x",
		Headers: []model.HTTPHeader{
			{Key: "Accept", Value: "application/json"},
			{Key: "Authorization", Value: "Bearer token"},
		},
	}

	item := httpRequestToPostmanItem(req)
	if len(item.Request.Header) != 2 {
		t.Fatalf("expected 2 headers, got %d", len(item.Request.Header))
	}
	if item.Request.Header[0].Key != "Accept" {
		t.Errorf("first header: got %q", item.Request.Header[0].Key)
	}
}

func TestHTTPRequestToPostmanItem_NoBody(t *testing.T) {
	req := model.HTTPRequest{Name: "Test", Method: "GET", URL: "http://x"}
	item := httpRequestToPostmanItem(req)
	if item.Request.Body != nil {
		t.Error("expected nil body for GET without body")
	}
}

func TestHoistAuth_AllSameBearer(t *testing.T) {
	files := []model.HTTPFile{
		{
			Path: "logs/activity.http",
			Requests: []model.HTTPRequest{
				{Name: "List", Method: "GET", URL: "http://x/list", Headers: []model.HTTPHeader{
					{Key: "Authorization", Value: "{{TOKEN}}"},
					{Key: "Content-Type", Value: "application/json"},
				}},
				{Name: "Export", Method: "POST", URL: "http://x/export", Headers: []model.HTTPHeader{
					{Key: "Authorization", Value: "{{TOKEN}}"},
				}},
			},
		},
		{
			Path: "logs/events.http",
			Requests: []model.HTTPRequest{
				{Name: "Get Events", Method: "GET", URL: "http://x/events", Headers: []model.HTTPHeader{
					{Key: "Authorization", Value: "{{TOKEN}}"},
				}},
			},
		},
	}

	col := HTTPFilesToCollection(files, "Test", nil)

	// "Logs" folder should have auth set
	logsFolder := col.Item[0]
	if logsFolder.Name != "Logs" {
		t.Fatalf("expected Logs folder, got %q", logsFolder.Name)
	}
	if logsFolder.Auth == nil {
		t.Fatal("expected auth on Logs folder")
	}
	if logsFolder.Auth.Type != "apikey" {
		t.Errorf("auth type: got %q, want %q", logsFolder.Auth.Type, "apikey")
	}

	// Descendant requests should NOT have Authorization header
	for _, subfolder := range logsFolder.Item {
		for _, reqItem := range subfolder.Item {
			if reqItem.Request == nil {
				continue
			}
			for _, h := range reqItem.Request.Header {
				if h.Key == "Authorization" {
					t.Errorf("request %q should not have Authorization header (should inherit)", reqItem.Name)
				}
			}
		}
	}
}

func TestHoistAuth_BearerToken(t *testing.T) {
	files := []model.HTTPFile{
		{
			Path: "api/users.http",
			Requests: []model.HTTPRequest{
				{Name: "List", Method: "GET", URL: "http://x", Headers: []model.HTTPHeader{
					{Key: "Authorization", Value: "Bearer my-token"},
				}},
				{Name: "Create", Method: "POST", URL: "http://x", Headers: []model.HTTPHeader{
					{Key: "Authorization", Value: "Bearer my-token"},
				}},
			},
		},
	}

	col := HTTPFilesToCollection(files, "Test", nil)
	folder := col.Item[0]
	if folder.Auth == nil {
		t.Fatal("expected auth on folder")
	}
	if folder.Auth.Type != "apikey" {
		t.Errorf("auth type: got %q, want %q", folder.Auth.Type, "apikey")
	}
	// Should have value="Bearer my-token", key="Authorization", in="header"
	foundValue := false
	for _, kv := range folder.Auth.APIKey {
		if kv.Key == "value" && kv.Value == "Bearer my-token" {
			foundValue = true
		}
	}
	if !foundValue {
		t.Errorf("expected apikey value 'Bearer my-token', got %+v", folder.Auth.APIKey)
	}
}

func TestHoistAuth_MixedAuth_NoHoist(t *testing.T) {
	files := []model.HTTPFile{
		{
			Path: "api/mixed.http",
			Requests: []model.HTTPRequest{
				{Name: "Public", Method: "GET", URL: "http://x", Headers: []model.HTTPHeader{
					{Key: "Authorization", Value: "Bearer token-a"},
				}},
				{Name: "Admin", Method: "GET", URL: "http://x", Headers: []model.HTTPHeader{
					{Key: "Authorization", Value: "Bearer token-b"},
				}},
			},
		},
	}

	col := HTTPFilesToCollection(files, "Test", nil)
	folder := col.Item[0]
	if folder.Auth != nil {
		t.Error("should NOT hoist auth when requests have different tokens")
	}

	// Requests should keep their headers
	for _, sub := range folder.Item {
		for _, req := range sub.Item {
			if req.Request == nil {
				continue
			}
			found := false
			for _, h := range req.Request.Header {
				if h.Key == "Authorization" {
					found = true
				}
			}
			if !found {
				t.Errorf("request %q should still have Authorization header", req.Name)
			}
		}
	}
}

func TestHoistAuth_NoAuth_NoHoist(t *testing.T) {
	files := []model.HTTPFile{
		{
			Path: "public/health.http",
			Requests: []model.HTTPRequest{
				{Name: "Health", Method: "GET", URL: "http://x/health"},
			},
		},
	}

	col := HTTPFilesToCollection(files, "Test", nil)
	if col.Item[0].Auth != nil {
		t.Error("should not set auth on folder when requests have no auth")
	}
}
