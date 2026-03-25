package postman

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/drossan/http2postman/internal/model"
)

type mockHTTP struct {
	responses []mockResponse
	requests  []*http.Request
	bodies    []string
	callIdx   int
}

type mockResponse struct {
	status int
	body   string
}

func (m *mockHTTP) Do(req *http.Request) (*http.Response, error) {
	m.requests = append(m.requests, req)
	if req.Body != nil {
		b, _ := io.ReadAll(req.Body)
		m.bodies = append(m.bodies, string(b))
	} else {
		m.bodies = append(m.bodies, "")
	}
	resp := m.responses[m.callIdx]
	m.callIdx++
	return &http.Response{
		StatusCode: resp.status,
		Body:       io.NopCloser(strings.NewReader(resp.body)),
	}, nil
}

func sampleCollection() *model.PostmanCollection {
	return &model.PostmanCollection{
		Info: model.PostmanInfo{
			Name:      "Test API",
			PostmanID: "fixed-uuid-123",
			Schema:    model.PostmanSchemaV210,
		},
		Item: []model.PostmanItem{
			{Name: "Get Users", Request: &model.PostmanReq{Method: "GET", URL: model.PostmanURL{Raw: "http://x"}}},
		},
	}
}

func TestListWorkspaces(t *testing.T) {
	mock := &mockHTTP{responses: []mockResponse{
		{200, `{"workspaces":[{"id":"ws-1","name":"My Workspace"},{"id":"ws-2","name":"Team"}]}`},
	}}
	client := NewClient("test-key", mock)

	workspaces, err := client.ListWorkspaces()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(workspaces) != 2 {
		t.Fatalf("expected 2 workspaces, got %d", len(workspaces))
	}
	if workspaces[0].Name != "My Workspace" {
		t.Errorf("name: got %q", workspaces[0].Name)
	}
}

func TestFindCollectionByName_Found(t *testing.T) {
	mock := &mockHTTP{responses: []mockResponse{
		{200, `{"collections":[{"id":"abc","uid":"123-abc","name":"Test API"}]}`},
	}}
	client := NewClient("test-key", mock)

	uid, err := client.FindCollectionByName("Test API", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uid != "123-abc" {
		t.Errorf("uid: got %q, want %q", uid, "123-abc")
	}
}

func TestFindCollectionByName_NotFound(t *testing.T) {
	mock := &mockHTTP{responses: []mockResponse{
		{200, `{"collections":[{"id":"abc","uid":"123-abc","name":"Other API"}]}`},
	}}
	client := NewClient("test-key", mock)

	uid, err := client.FindCollectionByName("Test API", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uid != "" {
		t.Errorf("expected empty uid, got %q", uid)
	}
}

func TestFindCollectionByName_WithWorkspace(t *testing.T) {
	mock := &mockHTTP{responses: []mockResponse{
		{200, `{"collections":[{"id":"abc","uid":"123-abc","name":"Test API"}]}`},
	}}
	client := NewClient("test-key", mock)

	client.FindCollectionByName("Test API", "ws-1")

	if !strings.Contains(mock.requests[0].URL.RawQuery, "workspace=ws-1") {
		t.Errorf("expected workspace param, got URL: %s", mock.requests[0].URL.String())
	}
}

func TestGetCollectionVersion(t *testing.T) {
	mock := &mockHTTP{responses: []mockResponse{
		{200, `{"collection":{"info":{"name":"Test API","version":"1.2.3"}}}`},
	}}
	client := NewClient("test-key", mock)

	version, err := client.GetCollectionVersion("123-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != "1.2.3" {
		t.Errorf("version: got %q, want %q", version, "1.2.3")
	}
}

func TestGetCollectionVersion_NoVersion(t *testing.T) {
	mock := &mockHTTP{responses: []mockResponse{
		{200, `{"collection":{"info":{"name":"Test API"}}}`},
	}}
	client := NewClient("test-key", mock)

	version, err := client.GetCollectionVersion("123-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != "" {
		t.Errorf("expected empty version, got %q", version)
	}
}

func TestCreateCollection_WithWorkspace(t *testing.T) {
	mock := &mockHTTP{responses: []mockResponse{
		{200, `{"collection":{"id":"new","uid":"123-new"}}`},
	}}
	client := NewClient("test-key", mock)

	uid, err := client.CreateCollection(sampleCollection(), "ws-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uid != "123-new" {
		t.Errorf("uid: got %q", uid)
	}
	if !strings.Contains(mock.requests[0].URL.RawQuery, "workspace=ws-1") {
		t.Errorf("expected workspace param, got URL: %s", mock.requests[0].URL.String())
	}
}

func TestUpdateCollection(t *testing.T) {
	mock := &mockHTTP{responses: []mockResponse{
		{200, `{"collection":{"id":"abc","uid":"123-abc"}}`},
	}}
	client := NewClient("test-key", mock)

	err := client.UpdateCollection("123-abc", sampleCollection())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.requests[0].Method != "PUT" {
		t.Errorf("method: got %q, want PUT", mock.requests[0].Method)
	}
}

func TestPushCollection_CreatesWithVersion100(t *testing.T) {
	mock := &mockHTTP{responses: []mockResponse{
		{200, `{"collections":[]}`},                             // FindByName: empty
		{200, `{"collection":{"id":"new","uid":"123-new-id"}}`}, // Create
	}}
	client := NewClient("test-key", mock)

	result, err := client.PushCollection(sampleCollection(), "ws-1", "minor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Created {
		t.Error("expected created=true")
	}
	if result.Version != "1.0.0" {
		t.Errorf("version: got %q, want %q", result.Version, "1.0.0")
	}

	// Verify the body has version 1.0.0
	var envelope collectionEnvelope
	json.Unmarshal([]byte(mock.bodies[1]), &envelope)
	if envelope.Collection.Info.Version != "1.0.0" {
		t.Errorf("body version: got %q", envelope.Collection.Info.Version)
	}
}

func TestPushCollection_UpdatesWithMinorBump(t *testing.T) {
	mock := &mockHTTP{responses: []mockResponse{
		{200, `{"collections":[{"id":"abc","uid":"123-abc","name":"Test API"}]}`}, // FindByName
		{200, `{"collection":{"info":{"name":"Test API","version":"1.2.0"}}}`},    // GetVersion
		{200, `{"collection":{"id":"abc","uid":"123-abc"}}`},                      // Update
	}}
	client := NewClient("test-key", mock)

	result, err := client.PushCollection(sampleCollection(), "ws-1", "minor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Created {
		t.Error("expected created=false")
	}
	if result.Version != "1.3.0" {
		t.Errorf("version: got %q, want %q", result.Version, "1.3.0")
	}
}

func TestPushCollection_UpdatesWithPatchBump(t *testing.T) {
	mock := &mockHTTP{responses: []mockResponse{
		{200, `{"collections":[{"id":"abc","uid":"123-abc","name":"Test API"}]}`},
		{200, `{"collection":{"info":{"name":"Test API","version":"1.2.0"}}}`},
		{200, `{"collection":{"id":"abc","uid":"123-abc"}}`},
	}}
	client := NewClient("test-key", mock)

	result, err := client.PushCollection(sampleCollection(), "ws-1", "patch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "1.2.1" {
		t.Errorf("version: got %q, want %q", result.Version, "1.2.1")
	}
}

func TestPushCollection_UpdatesWithMajorBump(t *testing.T) {
	mock := &mockHTTP{responses: []mockResponse{
		{200, `{"collections":[{"id":"abc","uid":"123-abc","name":"Test API"}]}`},
		{200, `{"collection":{"info":{"name":"Test API","version":"1.2.3"}}}`},
		{200, `{"collection":{"id":"abc","uid":"123-abc"}}`},
	}}
	client := NewClient("test-key", mock)

	result, err := client.PushCollection(sampleCollection(), "ws-1", "major")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "2.0.0" {
		t.Errorf("version: got %q, want %q", result.Version, "2.0.0")
	}
}

func TestApplyBump(t *testing.T) {
	tests := []struct {
		current  string
		bump     string
		expected string
	}{
		{"1.0.0", "minor", "1.1.0"},
		{"1.2.3", "patch", "1.2.4"},
		{"1.2.3", "major", "2.0.0"},
		{"", "minor", "1.0.0"},
		{"invalid", "minor", "1.0.0"},
		{"1.0.0", "", "1.1.0"}, // default is minor
	}
	for _, tt := range tests {
		t.Run(tt.current+"_"+tt.bump, func(t *testing.T) {
			got := applyBump(tt.current, tt.bump)
			if got != tt.expected {
				t.Errorf("applyBump(%q, %q) = %q, want %q", tt.current, tt.bump, got, tt.expected)
			}
		})
	}
}
