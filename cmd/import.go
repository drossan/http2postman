package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import a Postman collection to HTTP files",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file := args[0]
		if err := importPostmanCollection(file); err != nil {
			fmt.Println("Error:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}

func importPostmanCollection(file string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	var collection map[string]interface{}
	if err := json.Unmarshal(content, &collection); err != nil {
		return err
	}

	if items, ok := collection["item"].([]interface{}); ok {
		return createHTTPFiles(items, "http-requests", nil)
	}

	return fmt.Errorf("invalid collection format")
}

func createHTTPFiles(items []interface{}, basePath string, parentAuth map[string]interface{}) error {
	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := itemMap["name"].(string)
		groupPath := filepath.Join(basePath, formatFileName(name))

		var auth map[string]interface{}
		if a, ok := itemMap["auth"].(map[string]interface{}); ok {
			auth = a
		} else {
			auth = parentAuth
		}

		if subItems, ok := itemMap["item"].([]interface{}); ok {
			if err := os.MkdirAll(groupPath, os.ModePerm); err != nil {
				return err
			}
			if err := createHTTPFiles(subItems, groupPath, auth); err != nil {
				return err
			}
		} else {
			if err := createHTTPRequestFile(groupPath, itemMap, auth); err != nil {
				return err
			}
		}
	}

	return nil
}

func createHTTPRequestFile(filePath string, item map[string]interface{}, parentAuth map[string]interface{}) error {
	request, ok := item["request"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid request format")
	}

	method, _ := request["method"].(string)
	url, _ := extractURL(request["url"])
	headers, _ := request["header"].([]interface{})
	body, _ := request["body"].(map[string]interface{})

	// Add parent auth if not present in the request
	if _, ok := request["auth"]; !ok && parentAuth != nil {
		if bearerToken, found := extractBearerToken(parentAuth); found {
			headers = append(headers, map[string]interface{}{
				"key":   "Authorization",
				"value": "Bearer " + bearerToken,
			})
		}
	}

	var headerLines []string
	for _, header := range headers {
		headerMap, ok := header.(map[string]interface{})
		if !ok {
			continue
		}
		key, _ := headerMap["key"].(string)
		value, _ := headerMap["value"].(string)
		headerLines = append(headerLines, fmt.Sprintf("%s: %s", key, value))
	}

	var bodyContent string
	if body != nil && body["mode"] == "raw" {
		bodyContent, _ = body["raw"].(string)
	} else if body != nil && body["mode"] == "formdata" {
		formData, _ := body["formdata"].([]interface{})
		for _, field := range formData {
			fieldMap, ok := field.(map[string]interface{})
			if !ok {
				continue
			}
			key, _ := fieldMap["key"].(string)
			value, _ := fieldMap["value"].(string)
			if key != "" {
				bodyContent += fmt.Sprintf("%s: %s\n", key, value)
			}
		}
	}

	content := fmt.Sprintf("# %s\n%s %s\n%s\n\n%s\n",
		item["name"], method, url,
		strings.Join(headerLines, "\n"), bodyContent)

	fileName := filepath.Join(filePath + ".http")
	return os.WriteFile(fileName, []byte(content), 0644)
}

func extractBearerToken(auth map[string]interface{}) (string, bool) {
	if authType, ok := auth["type"].(string); ok && authType == "bearer" {
		if bearerList, ok := auth["bearer"].([]interface{}); ok {
			for _, b := range bearerList {
				bMap, ok := b.(map[string]interface{})
				if !ok {
					continue
				}
				if key, ok := bMap["key"].(string); ok && key == "token" {
					if value, ok := bMap["value"].(string); ok {
						return value, true
					}
				}
			}
		}
	}
	return "", false
}

func extractURL(url interface{}) (string, bool) {
	if urlMap, ok := url.(map[string]interface{}); ok {
		if raw, ok := urlMap["raw"].(string); ok {
			return raw, true
		}
	}
	return "", false
}

func formatFileName(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	return strings.ToLower(name)
}
