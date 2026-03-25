package converter

import (
	"path/filepath"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/drossan/http2postman/internal/model"
)

var titleCaser = cases.Title(language.Und)

// HTTPFilesToCollection converts parsed HTTP files into a Postman collection.
func HTTPFilesToCollection(files []model.HTTPFile, name string, version string, env *model.Environment) *model.PostmanCollection {
	displayName := name
	if version != "" {
		displayName = name + " v" + version
	}

	collection := &model.PostmanCollection{
		Info: model.PostmanInfo{
			Name:        displayName,
			PostmanID:   "generated-id",
			Description: "Generated from HTTP files",
			Schema:      model.PostmanSchemaV210,
			Version:     version,
		},
	}

	for _, file := range files {
		items := httpRequestsToPostmanItems(file.Requests)
		dirParts := strings.Split(filepath.Dir(file.Path), string(filepath.Separator))

		addToHierarchy(&collection.Item, dirParts, items)
	}

	// Hoist shared auth to parent folders so children use "Inherit auth from parent"
	hoistAuth(&collection.Item)

	if env != nil {
		collection.Variable = environmentToVars(*env)
	}

	return collection
}

func httpRequestsToPostmanItems(requests []model.HTTPRequest) []model.PostmanItem {
	items := make([]model.PostmanItem, 0, len(requests))
	for _, req := range requests {
		items = append(items, httpRequestToPostmanItem(req))
	}
	return items
}

func httpRequestToPostmanItem(req model.HTTPRequest) model.PostmanItem {
	item := model.PostmanItem{
		Name: req.Name,
		Request: &model.PostmanReq{
			Method: req.Method,
			URL:    model.PostmanURL{Raw: req.URL},
		},
	}

	for _, h := range req.Headers {
		item.Request.Header = append(item.Request.Header, model.PostmanHeader{
			Key:   h.Key,
			Value: h.Value,
		})
	}

	if req.Body != "" {
		item.Request.Body = &model.PostmanBody{
			Mode: "raw",
			Raw:  req.Body,
		}
	}

	return item
}

func addToHierarchy(items *[]model.PostmanItem, dirParts []string, requests []model.PostmanItem) {
	current := items

	for _, part := range dirParts {
		if part == "." {
			continue
		}
		formattedPart := formatGroupName(part)
		found := false
		for i := range *current {
			if (*current)[i].Name == formattedPart {
				current = &(*current)[i].Item
				found = true
				break
			}
		}
		if !found {
			*current = append(*current, model.PostmanItem{
				Name: formattedPart,
			})
			current = &(*current)[len(*current)-1].Item
		}
	}

	*current = append(*current, requests...)
}

// FormatGroupName formats a name for display (exported for testing).
func FormatGroupName(name string) string {
	return formatGroupName(name)
}

func formatGroupName(name string) string {
	name = strings.TrimSuffix(name, ".http")
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")
	return titleCaser.String(name)
}

// hoistAuth recursively checks folders: if all descendant requests share the
// same Authorization header, it sets auth at the folder level and removes the
// header from individual requests so they use "Inherit auth from parent".
func hoistAuth(items *[]model.PostmanItem) {
	for i := range *items {
		item := &(*items)[i]
		if !item.IsFolder() || len(item.Item) == 0 {
			continue
		}

		// Recurse into subfolders first (bottom-up)
		hoistAuth(&item.Item)

		// Collect all auth values from descendant requests
		authValue := collectCommonAuth(item.Item)
		if authValue == "" {
			continue
		}

		// Set auth at folder level using apikey type with an Authorization header.
		// This is the most flexible format — works for Bearer, API keys, and variables.
		item.Auth = &model.PostmanAuth{
			Type: "apikey",
			APIKey: []model.PostmanKV{
				{Key: "key", Value: "Authorization"},
				{Key: "value", Value: authValue},
				{Key: "in", Value: "header"},
			},
		}

		// Remove Authorization header from all descendant requests
		removeAuthHeader(&item.Item)
	}
}

// collectCommonAuth returns the common Authorization header value if ALL
// descendant requests share the same one. Returns "" if they differ or none have it.
func collectCommonAuth(items []model.PostmanItem) string {
	var commonAuth string
	first := true

	for _, item := range items {
		if item.IsFolder() {
			// Check subfolder: if it already has auth set, use that value
			if item.Auth != nil {
				val := authToHeaderValue(item.Auth)
				if first {
					commonAuth = val
					first = false
				} else if val != commonAuth {
					return ""
				}
			} else {
				// Recurse into subfolder without its own auth
				sub := collectCommonAuth(item.Item)
				if sub == "" {
					return ""
				}
				if first {
					commonAuth = sub
					first = false
				} else if sub != commonAuth {
					return ""
				}
			}
		} else if item.Request != nil {
			val := getAuthHeaderValue(item.Request.Header)
			if val == "" {
				// Request without auth — can't hoist
				return ""
			}
			if first {
				commonAuth = val
				first = false
			} else if val != commonAuth {
				return ""
			}
		}
	}

	return commonAuth
}

// getAuthHeaderValue returns the value of the Authorization header, or "".
func getAuthHeaderValue(headers []model.PostmanHeader) string {
	for _, h := range headers {
		if strings.EqualFold(h.Key, "Authorization") {
			return h.Value
		}
	}
	return ""
}

// authToHeaderValue converts a PostmanAuth back to its header value representation.
func authToHeaderValue(auth *model.PostmanAuth) string {
	if auth.Type == "bearer" {
		for _, kv := range auth.Bearer {
			if kv.Key == "token" {
				return "Bearer " + kv.Value
			}
		}
	}
	if auth.Type == "apikey" {
		for _, kv := range auth.APIKey {
			if kv.Key == "value" {
				return kv.Value
			}
		}
	}
	return ""
}

// removeAuthHeader removes Authorization headers from all descendant requests
// and clears auth from subfolders (they will inherit from parent).
func removeAuthHeader(items *[]model.PostmanItem) {
	for i := range *items {
		item := &(*items)[i]
		if item.IsFolder() {
			// Remove subfolder's own auth — it will inherit from parent
			item.Auth = nil
			removeAuthHeader(&item.Item)
		} else if item.Request != nil {
			var filtered []model.PostmanHeader
			for _, h := range item.Request.Header {
				if !strings.EqualFold(h.Key, "Authorization") {
					filtered = append(filtered, h)
				}
			}
			item.Request.Header = filtered
		}
	}
}

func environmentToVars(env model.Environment) []model.PostmanVar {
	var vars []model.PostmanVar
	for _, envVars := range env {
		for key, value := range envVars {
			vars = append(vars, model.PostmanVar{
				Key:   key,
				Value: value,
				Type:  "string",
			})
		}
	}
	return vars
}
