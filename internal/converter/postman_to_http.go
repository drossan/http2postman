package converter

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/drossan/http2postman/internal/model"
)

// CollectionToHTTPFiles converts a Postman collection into HTTP files.
func CollectionToHTTPFiles(collection *model.PostmanCollection) []model.HTTPFile {
	return itemsToHTTPFiles(collection.Item, "", nil)
}

func itemsToHTTPFiles(items []model.PostmanItem, basePath string, parentAuth *model.PostmanAuth) []model.HTTPFile {
	var files []model.HTTPFile

	for _, item := range items {
		auth := item.Auth
		if auth == nil {
			auth = parentAuth
		}

		if item.IsFolder() {
			dirName := FormatFileName(item.Name)
			subPath := filepath.Join(basePath, dirName)
			files = append(files, itemsToHTTPFiles(item.Item, subPath, auth)...)
		} else {
			httpFile := postmanItemToHTTPFile(item, basePath, auth)
			files = append(files, httpFile)
		}
	}

	return files
}

func postmanItemToHTTPFile(item model.PostmanItem, basePath string, parentAuth *model.PostmanAuth) model.HTTPFile {
	req := item.Request
	httpReq := model.HTTPRequest{
		Name:   item.Name,
		Method: req.Method,
		URL:    req.URL.Raw,
	}

	for _, h := range req.Header {
		httpReq.Headers = append(httpReq.Headers, model.HTTPHeader{
			Key:   h.Key,
			Value: h.Value,
		})
	}

	// Inherit auth from parent if request has no auth
	if req.Auth == nil && parentAuth != nil {
		if token, ok := ExtractBearerToken(parentAuth); ok {
			httpReq.Headers = append(httpReq.Headers, model.HTTPHeader{
				Key:   "Authorization",
				Value: fmt.Sprintf("Bearer %s", token),
			})
		}
	}

	if req.Body != nil {
		switch req.Body.Mode {
		case "raw":
			httpReq.Body = req.Body.Raw
		case "formdata":
			var lines []string
			for _, field := range req.Body.FormData {
				if field.Key != "" {
					lines = append(lines, fmt.Sprintf("%s: %s", field.Key, field.Value))
				}
			}
			httpReq.Body = strings.Join(lines, "\n")
		}
	}

	fileName := FormatFileName(item.Name)
	return model.HTTPFile{
		Path:     filepath.Join(basePath, fileName+".http"),
		Requests: []model.HTTPRequest{httpReq},
	}
}

// ExtractBearerToken extracts the bearer token from a PostmanAuth.
func ExtractBearerToken(auth *model.PostmanAuth) (string, bool) {
	if auth == nil || auth.Type != "bearer" {
		return "", false
	}
	for _, kv := range auth.Bearer {
		if kv.Key == "token" {
			return kv.Value, true
		}
	}
	return "", false
}

// FormatFileName sanitizes a name for use as a file/directory name.
func FormatFileName(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	return strings.ToLower(name)
}
