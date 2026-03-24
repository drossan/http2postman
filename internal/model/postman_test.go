package model

import (
	"encoding/json"
	"testing"
)

func TestPostmanItem_IsFolder_WithSubItems(t *testing.T) {
	item := PostmanItem{
		Name: "My Folder",
		Item: []PostmanItem{{Name: "child"}},
	}
	if !item.IsFolder() {
		t.Error("expected IsFolder() to return true for item with sub-items and no request")
	}
}

func TestPostmanItem_IsFolder_WithRequest(t *testing.T) {
	item := PostmanItem{
		Name:    "My Request",
		Request: &PostmanReq{Method: "GET"},
	}
	if item.IsFolder() {
		t.Error("expected IsFolder() to return false for item with request")
	}
}

func TestPostmanItem_IsFolder_Empty(t *testing.T) {
	item := PostmanItem{Name: "Empty"}
	if !item.IsFolder() {
		t.Error("expected IsFolder() to return true for item with no request (even if no sub-items)")
	}
}

func TestPostmanURL_UnmarshalJSON_StringURL(t *testing.T) {
	data := []byte(`"https://example.com/api"`)
	var url PostmanURL
	if err := json.Unmarshal(data, &url); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url.Raw != "https://example.com/api" {
		t.Errorf("expected Raw = %q, got %q", "https://example.com/api", url.Raw)
	}
}

func TestPostmanURL_UnmarshalJSON_ObjectURL(t *testing.T) {
	data := []byte(`{"raw": "https://example.com/api"}`)
	var url PostmanURL
	if err := json.Unmarshal(data, &url); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url.Raw != "https://example.com/api" {
		t.Errorf("expected Raw = %q, got %q", "https://example.com/api", url.Raw)
	}
}

func TestPostmanURL_UnmarshalJSON_InvalidJSON(t *testing.T) {
	data := []byte(`[invalid]`)
	var url PostmanURL
	if err := json.Unmarshal(data, &url); err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestPostmanCollection_MarshalRoundTrip(t *testing.T) {
	original := PostmanCollection{
		Info: PostmanInfo{
			Name:        "Test Collection",
			PostmanID:   "test-id",
			Description: "A test",
			Schema:      PostmanSchemaV210,
		},
		Item: []PostmanItem{
			{
				Name: "Get Users",
				Request: &PostmanReq{
					Method: "GET",
					URL:    PostmanURL{Raw: "https://api.example.com/users"},
					Header: []PostmanHeader{{Key: "Authorization", Value: "Bearer token"}},
				},
			},
			{
				Name: "Auth Folder",
				Item: []PostmanItem{
					{
						Name: "Login",
						Request: &PostmanReq{
							Method: "POST",
							URL:    PostmanURL{Raw: "https://api.example.com/login"},
							Body:   &PostmanBody{Mode: "raw", Raw: `{"user":"test"}`},
						},
					},
				},
				Auth: &PostmanAuth{
					Type:   "bearer",
					Bearer: []PostmanKV{{Key: "token", Value: "abc123"}},
				},
			},
		},
		Variable: []PostmanVar{{Key: "host", Value: "https://api.example.com", Type: "string"}},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded PostmanCollection
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.Info.Name != original.Info.Name {
		t.Errorf("Info.Name: got %q, want %q", decoded.Info.Name, original.Info.Name)
	}
	if len(decoded.Item) != 2 {
		t.Fatalf("expected 2 items, got %d", len(decoded.Item))
	}
	if decoded.Item[0].Request.Method != "GET" {
		t.Errorf("first item method: got %q, want %q", decoded.Item[0].Request.Method, "GET")
	}
	if decoded.Item[1].IsFolder() != true {
		t.Error("second item should be a folder")
	}
	if len(decoded.Item[1].Item) != 1 {
		t.Fatalf("folder should have 1 sub-item, got %d", len(decoded.Item[1].Item))
	}
	if decoded.Item[1].Auth.Type != "bearer" {
		t.Errorf("folder auth type: got %q, want %q", decoded.Item[1].Auth.Type, "bearer")
	}
	if len(decoded.Variable) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(decoded.Variable))
	}
	if decoded.Variable[0].Key != "host" {
		t.Errorf("variable key: got %q, want %q", decoded.Variable[0].Key, "host")
	}
}
