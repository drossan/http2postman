package writer

import (
	"fmt"
	"strings"

	"github.com/drossan/http2postman/internal/fs"
	"github.com/drossan/http2postman/internal/model"
)

// HTTPFileWriter writes HTTP files to disk.
type HTTPFileWriter struct {
	fs fs.FileSystem
}

// NewHTTPFileWriter creates a new writer with the given filesystem.
func NewHTTPFileWriter(fsys fs.FileSystem) *HTTPFileWriter {
	return &HTTPFileWriter{fs: fsys}
}

// Write writes all HTTP files under baseDir.
// If force is false and a file already exists, it returns an error.
func (w *HTTPFileWriter) Write(files []model.HTTPFile, baseDir string, force bool) error {
	for _, file := range files {
		path := file.Path
		if baseDir != "" {
			path = baseDir + "/" + file.Path
		}

		if !force && w.fs.FileExists(path) {
			return fmt.Errorf("output file %s already exists, use --force to overwrite", path)
		}

		content := FormatHTTPFile(file)
		if err := w.fs.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing HTTP file %s: %w", path, err)
		}
	}
	return nil
}

// FormatHTTPFile formats an HTTPFile into the .http text format.
func FormatHTTPFile(file model.HTTPFile) string {
	var sections []string
	for _, req := range file.Requests {
		sections = append(sections, FormatHTTPRequest(req))
	}
	return strings.Join(sections, "\n###\n\n")
}

// FormatHTTPRequest formats a single HTTP request in .http format.
func FormatHTTPRequest(req model.HTTPRequest) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n", req.Name))
	sb.WriteString(fmt.Sprintf("%s %s\n", req.Method, req.URL))

	for _, h := range req.Headers {
		sb.WriteString(fmt.Sprintf("%s: %s\n", h.Key, h.Value))
	}

	if req.Body != "" {
		sb.WriteString("\n")
		sb.WriteString(req.Body)
		sb.WriteString("\n")
	}

	return sb.String()
}
