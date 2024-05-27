package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export [directory]",
	Short: "Export HTTP requests to a Postman collection",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := args[0]
		collectionName := promptCollectionName()
		if err := processDirectory(dir, collectionName); err != nil {
			fmt.Println("Error:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}

func promptCollectionName() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the name for the Postman collection: ")
	name, _ := reader.ReadString('\n')
	return strings.TrimSpace(name)
}

func processDirectory(dir string, collectionName string) error {
	collection := map[string]interface{}{
		"info": map[string]interface{}{
			"name":        collectionName,
			"_postman_id": "unique-id",
			"description": "Generated from HTTP files",
			"schema":      "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
		"item":     []interface{}{},
		"variable": []map[string]interface{}{},
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".http") {
			items, err := processHTTPFile(path)
			if err != nil {
				return err
			}
			relativePath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			dirParts := strings.Split(filepath.Dir(relativePath), string(filepath.Separator))
			groupName := formatGroupName(filepath.Base(path))

			addToCollection(collection, dirParts, groupName, items)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Add variables from http-client.env.json if it exists
	envFilePath, err := findEnvFile(dir)
	if err == nil && envFilePath != "" {
		if err := addVariablesFromEnvFile(envFilePath, &collection, collectionName); err != nil {
			return err
		}
	}

	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("import_postman_collection.json", data, 0644)
}

func addToCollection(collection map[string]interface{}, dirParts []string, groupName string, items []map[string]interface{}) {
	currentItems := collection["item"].([]interface{})
	var parentGroup map[string]interface{}

	for _, part := range dirParts {
		formattedPart := formatGroupName(part)
		found := false
		for _, item := range currentItems {
			group := item.(map[string]interface{})
			if group["name"] == formattedPart {
				parentGroup = group
				currentItems = group["item"].([]interface{})
				found = true
				break
			}
		}
		if !found {
			newGroup := map[string]interface{}{
				"name": formattedPart,
				"item": []interface{}{},
			}
			currentItems = append(currentItems, newGroup)
			if parentGroup != nil {
				parentGroup["item"] = currentItems
			} else {
				collection["item"] = currentItems
			}
			parentGroup = newGroup
			currentItems = newGroup["item"].([]interface{})
		}
	}

	newItem := map[string]interface{}{
		"name": groupName,
		"item": items,
	}

	currentItems = append(currentItems, newItem)
	if parentGroup != nil {
		parentGroup["item"] = currentItems
	} else {
		collection["item"] = currentItems
	}
}

func findEnvFile(dir string) (string, error) {
	for {
		envFilePath := filepath.Join(dir, "http-client.env.json")
		if _, err := os.Stat(envFilePath); err == nil {
			return envFilePath, nil
		}
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break
		}
		dir = parentDir
	}
	return "", fmt.Errorf("http-client.env.json not found")
}

func formatGroupName(name string) string {
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")
	return strings.Title(name)
}

func processHTTPFile(path string) ([]map[string]interface{}, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	sections := strings.Split(string(content), "###")
	var items []map[string]interface{}
	for _, section := range sections {
		lines := strings.Split(strings.TrimSpace(section), "\n")
		if len(lines) == 0 {
			continue
		}
		if len(lines) < 2 {
			fmt.Printf("Invalid HTTP request format in file: %s\n", path)
			continue
		}
		urlLine := strings.SplitN(lines[1], " ", 2)
		if len(urlLine) < 2 {
			fmt.Printf("Invalid URL line format in file: %s\n", path)
			continue
		}
		headers, body := processHeadersAndBody(lines[2:])

		item := map[string]interface{}{
			"name": strings.TrimPrefix(lines[0], "# "),
			"request": map[string]interface{}{
				"method": urlLine[0],
				"url":    urlLine[1],
				"header": headers,
			},
		}

		if body != "" {
			item["request"].(map[string]interface{})["body"] = map[string]interface{}{
				"mode": "raw",
				"raw":  body,
			}
		}

		items = append(items, item)
	}
	return items, nil
}

func processHeadersAndBody(lines []string) ([]map[string]string, string) {
	var headers []map[string]string
	var bodyLines []string
	headersEnded := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			headersEnded = true
			continue
		}

		if headersEnded {
			bodyLines = append(bodyLines, line)
		} else {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				headers = append(headers, map[string]string{"key": strings.TrimSpace(parts[0]), "value": strings.TrimSpace(parts[1])})
			} else {
				fmt.Printf("Invalid header format: %s\n", line)
			}
		}
	}

	body := strings.Join(bodyLines, "\n")
	return headers, body
}

func addVariablesFromEnvFile(envFilePath string, collection *map[string]interface{}, collectionName string) error {
	content, err := os.ReadFile(envFilePath)
	if err != nil {
		return err
	}

	var envData map[string]map[string]string
	if err := json.Unmarshal(content, &envData); err != nil {
		return err
	}

	for _, envVariables := range envData {
		for key, value := range envVariables {
			variable := map[string]interface{}{
				"key":   key,
				"value": value,
				"type":  "string",
			}
			(*collection)["variable"] = append((*collection)["variable"].([]map[string]interface{}), variable)
		}
	}

	return nil
}
