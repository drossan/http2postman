package apidog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/drossan/http2postman/internal/model"
)

const (
	baseURL    = "https://api.apidog.com"
	apiVersion = "2024-03-28"
)

// Client interacts with the Apidog API.
type Client struct {
	token      string
	httpClient HTTPDoer
}

// HTTPDoer abstracts HTTP calls for testing.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewClient creates an Apidog API client with the given access token.
func NewClient(token string, httpClient HTTPDoer) *Client {
	return &Client{token: token, httpClient: httpClient}
}

// Project represents an Apidog project.
type Project struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type listProjectsResponse struct {
	Data []Project `json:"data"`
}

type importRequest struct {
	Input   string        `json:"input"`
	Options importOptions `json:"options"`
}

type importOptions struct {
	EndpointOverwriteBehavior string `json:"endpointOverwriteBehavior"`
}

type importResponse struct {
	Data importResult `json:"data"`
}

type importResult struct {
	Endpoints importCounts `json:"endpoints"`
	Folders   importCounts `json:"endpointFolders"`
}

// ImportCounts holds create/update/skip counts from an import.
type importCounts struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
	Ignored int `json:"ignored"`
	Failed  int `json:"failed"`
}

// ImportResult holds the result of an import operation.
type ImportResult struct {
	EndpointsCreated int
	EndpointsUpdated int
	FoldersCreated   int
	FoldersUpdated   int
}

// ListProjects returns all projects accessible with the token.
func (c *Client) ListProjects() ([]Project, error) {
	req, err := http.NewRequest("GET", baseURL+"/v1/projects", nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("listing projects: status %d: %s", resp.StatusCode, respBody)
	}

	var result listProjectsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding projects: %w", err)
	}
	return result.Data, nil
}

// PushCollection imports a Postman collection into an Apidog project.
// It uses OVERWRITE_EXISTING to update endpoints that already exist.
func (c *Client) PushCollection(projectID int, collection *model.PostmanCollection) (*ImportResult, error) {
	collectionJSON, err := json.Marshal(collection)
	if err != nil {
		return nil, fmt.Errorf("marshaling collection: %w", err)
	}

	body, err := json.Marshal(importRequest{
		Input: string(collectionJSON),
		Options: importOptions{
			EndpointOverwriteBehavior: "OVERWRITE_EXISTING",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling import request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/projects/%d/import-postman-collection", baseURL, projectID)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("importing collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("importing collection: status %d: %s", resp.StatusCode, respBody)
	}

	var result importResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding import response: %w", err)
	}

	return &ImportResult{
		EndpointsCreated: result.Data.Endpoints.Created,
		EndpointsUpdated: result.Data.Endpoints.Updated,
		FoldersCreated:   result.Data.Folders.Created,
		FoldersUpdated:   result.Data.Folders.Updated,
	}, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("X-Apidog-Api-Version", apiVersion)
}
