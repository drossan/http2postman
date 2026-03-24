package parser

import (
	"encoding/json"
	"fmt"

	"github.com/drossan/http2postman/internal/fs"
	"github.com/drossan/http2postman/internal/model"
)

// ParsePostmanCollection parses raw JSON bytes into a PostmanCollection.
func ParsePostmanCollection(data []byte) (*model.PostmanCollection, error) {
	var collection model.PostmanCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, fmt.Errorf("parsing Postman collection JSON: %w", err)
	}
	if collection.Item == nil {
		return nil, fmt.Errorf("%w: missing 'item' field", model.ErrInvalidCollection)
	}
	return &collection, nil
}

// ParsePostmanCollectionFromFile reads and parses a Postman collection JSON file.
func ParsePostmanCollectionFromFile(fsys fs.FileSystem, path string) (*model.PostmanCollection, error) {
	data, err := fsys.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading collection file %s: %w", path, err)
	}
	return ParsePostmanCollection(data)
}
