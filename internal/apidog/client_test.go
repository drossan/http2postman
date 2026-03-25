package apidog

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
			Name:   "Test API",
			Schema: model.PostmanSchemaV210,
		},
		Item: []model.PostmanItem{
			{Name: "Get Users", Request: &model.PostmanReq{Method: "GET", URL: model.PostmanURL{Raw: "http://x"}}},
		},
	}
}

func TestListProjects(t *testing.T) {
	mock := &mockHTTP{responses: []mockResponse{
		{200, `{"data":[{"id":123,"name":"My Project"},{"id":456,"name":"Other"}]}`},
	}}
	client := NewClient("test-token", mock)

	projects, err := client.ListProjects()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}
	if projects[0].Name != "My Project" {
		t.Errorf("name: got %q", projects[0].Name)
	}
	if projects[0].ID != 123 {
		t.Errorf("id: got %d", projects[0].ID)
	}
}

func TestListProjects_AuthHeaders(t *testing.T) {
	mock := &mockHTTP{responses: []mockResponse{
		{200, `{"data":[]}`},
	}}
	client := NewClient("my-token", mock)
	client.ListProjects()

	req := mock.requests[0]
	if req.Header.Get("Authorization") != "Bearer my-token" {
		t.Errorf("auth header: got %q", req.Header.Get("Authorization"))
	}
	if req.Header.Get("X-Apidog-Api-Version") != "2024-03-28" {
		t.Errorf("api version header: got %q", req.Header.Get("X-Apidog-Api-Version"))
	}
}

func TestListProjects_Error(t *testing.T) {
	mock := &mockHTTP{responses: []mockResponse{
		{401, `{"error":"unauthorized"}`},
	}}
	client := NewClient("bad-token", mock)

	_, err := client.ListProjects()
	if err == nil {
		t.Fatal("expected error for 401")
	}
}

func TestPushCollection(t *testing.T) {
	mock := &mockHTTP{responses: []mockResponse{
		{200, `{"data":{"endpoints":{"created":3,"updated":1,"ignored":0,"failed":0},"endpointFolders":{"created":2,"updated":0,"ignored":0,"failed":0}}}`},
	}}
	client := NewClient("test-token", mock)

	result, err := client.PushCollection(123, sampleCollection())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.EndpointsCreated != 3 {
		t.Errorf("endpoints created: got %d, want 3", result.EndpointsCreated)
	}
	if result.EndpointsUpdated != 1 {
		t.Errorf("endpoints updated: got %d, want 1", result.EndpointsUpdated)
	}
	if result.FoldersCreated != 2 {
		t.Errorf("folders created: got %d, want 2", result.FoldersCreated)
	}
}

func TestPushCollection_URL(t *testing.T) {
	mock := &mockHTTP{responses: []mockResponse{
		{200, `{"data":{"endpoints":{},"endpointFolders":{}}}`},
	}}
	client := NewClient("test-token", mock)
	client.PushCollection(456, sampleCollection())

	if !strings.Contains(mock.requests[0].URL.Path, "/v1/projects/456/import-postman-collection") {
		t.Errorf("url: got %q", mock.requests[0].URL.Path)
	}
}

func TestPushCollection_BodyFormat(t *testing.T) {
	mock := &mockHTTP{responses: []mockResponse{
		{200, `{"data":{"endpoints":{},"endpointFolders":{}}}`},
	}}
	client := NewClient("test-token", mock)
	client.PushCollection(123, sampleCollection())

	var req importRequest
	if err := json.Unmarshal([]byte(mock.bodies[0]), &req); err != nil {
		t.Fatalf("invalid body: %v", err)
	}
	if req.Options.EndpointOverwriteBehavior != "OVERWRITE_EXISTING" {
		t.Errorf("overwrite behavior: got %q", req.Options.EndpointOverwriteBehavior)
	}
	// Input should be a stringified JSON of the collection
	var col model.PostmanCollection
	if err := json.Unmarshal([]byte(req.Input), &col); err != nil {
		t.Fatalf("input is not valid collection JSON: %v", err)
	}
	if col.Info.Name != "Test API" {
		t.Errorf("collection name in input: got %q", col.Info.Name)
	}
}

func TestPushCollection_Error(t *testing.T) {
	mock := &mockHTTP{responses: []mockResponse{
		{500, `{"error":"internal"}`},
	}}
	client := NewClient("test-token", mock)

	_, err := client.PushCollection(123, sampleCollection())
	if err == nil {
		t.Fatal("expected error for 500")
	}
}
