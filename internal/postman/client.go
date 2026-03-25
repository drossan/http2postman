package postman

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/drossan/http2postman/internal/model"
)

const baseURL = "https://api.getpostman.com"

// Client interacts with the Postman API to manage collections.
type Client struct {
	apiKey     string
	httpClient HTTPDoer
}

// HTTPDoer abstracts HTTP calls for testing.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewClient creates a Postman API client with the given API key.
func NewClient(apiKey string, httpClient HTTPDoer) *Client {
	return &Client{apiKey: apiKey, httpClient: httpClient}
}

// Workspace represents a Postman workspace.
type Workspace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type collectionEnvelope struct {
	Collection *model.PostmanCollection `json:"collection"`
}

type listResponse struct {
	Collections []collectionEntry `json:"collections"`
}

type collectionEntry struct {
	ID   string `json:"id"`
	UID  string `json:"uid"`
	Name string `json:"name"`
}

type createResponse struct {
	Collection struct {
		ID  string `json:"id"`
		UID string `json:"uid"`
	} `json:"collection"`
}

type workspacesResponse struct {
	Workspaces []Workspace `json:"workspaces"`
}

// ListWorkspaces returns all workspaces accessible with the API key.
func (c *Client) ListWorkspaces() ([]Workspace, error) {
	req, err := http.NewRequest("GET", baseURL+"/workspaces", nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("listing workspaces: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("listing workspaces: status %d", resp.StatusCode)
	}

	var result workspacesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding workspaces: %w", err)
	}
	return result.Workspaces, nil
}

// FindCollectionByName searches for a collection by name in a workspace.
// If workspaceID is empty, searches the default workspace.
// Returns the UID if found, or empty string if not.
func (c *Client) FindCollectionByName(name string, workspaceID string) (string, error) {
	url := baseURL + "/collections"
	if workspaceID != "" {
		url += "?workspace=" + workspaceID
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("building request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("listing collections: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("listing collections: status %d", resp.StatusCode)
	}

	var result listResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding collection list: %w", err)
	}

	for _, col := range result.Collections {
		if col.Name == name {
			return col.UID, nil
		}
	}
	return "", nil
}

// CreateCollection creates a new collection and returns its UID.
// If workspaceID is empty, creates in the default workspace.
func (c *Client) CreateCollection(collection *model.PostmanCollection, workspaceID string) (string, error) {
	body, err := json.Marshal(collectionEnvelope{Collection: collection})
	if err != nil {
		return "", fmt.Errorf("marshaling collection: %w", err)
	}

	url := baseURL + "/collections"
	if workspaceID != "" {
		url += "?workspace=" + workspaceID
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("building request: %w", err)
	}
	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("creating collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("creating collection: status %d: %s", resp.StatusCode, respBody)
	}

	var result createResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding create response: %w", err)
	}

	return result.Collection.UID, nil
}

// UpdateCollection replaces an existing collection by UID.
func (c *Client) UpdateCollection(uid string, collection *model.PostmanCollection) error {
	body, err := json.Marshal(collectionEnvelope{Collection: collection})
	if err != nil {
		return fmt.Errorf("marshaling collection: %w", err)
	}

	req, err := http.NewRequest("PUT", baseURL+"/collections/"+uid, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("updating collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("updating collection: status %d: %s", resp.StatusCode, respBody)
	}

	return nil
}

// GetCollectionVersion fetches a collection by UID and returns its version.
// Returns empty string if the collection has no version field.
func (c *Client) GetCollectionVersion(uid string) (string, error) {
	req, err := http.NewRequest("GET", baseURL+"/collections/"+uid, nil)
	if err != nil {
		return "", fmt.Errorf("building request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetching collection: status %d", resp.StatusCode)
	}

	var result collectionEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding collection: %w", err)
	}

	return result.Collection.Info.Version, nil
}

// PushResult contains the result of a push operation.
type PushResult struct {
	UID     string
	Created bool
	Version string
}

// PushCollection creates or updates a collection in a Postman workspace.
// It searches by name — if found, reads the current version from Postman,
// applies the bump, and updates. If not found, creates with version 1.0.0.
func (c *Client) PushCollection(collection *model.PostmanCollection, workspaceID string, bump string) (*PushResult, error) {
	uid, err := c.FindCollectionByName(collection.Info.Name, workspaceID)
	if err != nil {
		return nil, err
	}

	if uid != "" {
		// Read current version from Postman
		currentVersion, err := c.GetCollectionVersion(uid)
		if err != nil {
			return nil, fmt.Errorf("reading current version: %w", err)
		}

		newVersion := applyBump(currentVersion, bump)
		collection.Info.Version = newVersion

		if err := c.UpdateCollection(uid, collection); err != nil {
			return nil, err
		}
		return &PushResult{UID: uid, Created: false, Version: newVersion}, nil
	}

	collection.Info.Version = "1.0.0"
	uid, err = c.CreateCollection(collection, workspaceID)
	if err != nil {
		return nil, err
	}
	return &PushResult{UID: uid, Created: true, Version: "1.0.0"}, nil
}

// applyBump increments the version string according to the bump type.
// Returns "1.0.0" if the current version is empty or invalid.
func applyBump(current string, bump string) string {
	if current == "" {
		return "1.0.0"
	}
	parts := strings.Split(current, ".")
	if len(parts) != 3 {
		return "1.0.0"
	}
	major, err1 := strconv.Atoi(parts[0])
	minor, err2 := strconv.Atoi(parts[1])
	patch, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return "1.0.0"
	}

	switch bump {
	case "major":
		major++
		minor = 0
		patch = 0
	case "patch":
		patch++
	default: // "minor"
		minor++
		patch = 0
	}
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("X-API-Key", c.apiKey)
}
