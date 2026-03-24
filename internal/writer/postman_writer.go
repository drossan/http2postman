package writer

import (
	"encoding/json"
	"fmt"

	"github.com/drossan/http2postman/internal/fs"
	"github.com/drossan/http2postman/internal/model"
)

// PostmanWriter writes Postman collections to JSON files.
type PostmanWriter struct {
	fs fs.FileSystem
}

// NewPostmanWriter creates a new writer with the given filesystem.
func NewPostmanWriter(fsys fs.FileSystem) *PostmanWriter {
	return &PostmanWriter{fs: fsys}
}

// Write marshals the collection to JSON and writes it to outputPath.
// If force is false and the file already exists, it returns an error.
func (w *PostmanWriter) Write(collection *model.PostmanCollection, outputPath string, force bool) error {
	if !force && w.fs.FileExists(outputPath) {
		return fmt.Errorf("output file %s already exists, use --force to overwrite", outputPath)
	}

	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling collection to JSON: %w", err)
	}

	if err := w.fs.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("writing collection to %s: %w", outputPath, err)
	}

	return nil
}
