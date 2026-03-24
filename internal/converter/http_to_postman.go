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
func HTTPFilesToCollection(files []model.HTTPFile, name string, env *model.Environment) *model.PostmanCollection {
	collection := &model.PostmanCollection{
		Info: model.PostmanInfo{
			Name:        name,
			PostmanID:   "generated-id",
			Description: "Generated from HTTP files",
			Schema:      model.PostmanSchemaV210,
		},
	}

	for _, file := range files {
		items := httpRequestsToPostmanItems(file.Requests)
		groupName := formatGroupName(filepath.Base(file.Path))
		dirParts := strings.Split(filepath.Dir(file.Path), string(filepath.Separator))

		addToHierarchy(&collection.Item, dirParts, groupName, items)
	}

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

func addToHierarchy(items *[]model.PostmanItem, dirParts []string, groupName string, requests []model.PostmanItem) {
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

	fileGroup := model.PostmanItem{
		Name: groupName,
		Item: requests,
	}
	*current = append(*current, fileGroup)
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
