package parser

import (
	"fmt"
	ifs "io/fs"
	"path/filepath"
	"strings"

	"github.com/drossan/http2postman/internal/fs"
	"github.com/drossan/http2postman/internal/model"
)

// HTTPFileParser parses .http files using an injected filesystem.
type HTTPFileParser struct {
	fs fs.FileSystem
}

// NewHTTPFileParser creates a new parser with the given filesystem.
func NewHTTPFileParser(fsys fs.FileSystem) *HTTPFileParser {
	return &HTTPFileParser{fs: fsys}
}

// ParseFile reads and parses a single .http file.
func (p *HTTPFileParser) ParseFile(path string) (*model.HTTPFile, error) {
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}
	content, err := p.fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", path, err)
	}
	requests, err := ParseHTTPContent(string(content))
	if err != nil {
		return nil, fmt.Errorf("parsing file %s: %w", path, err)
	}
	return &model.HTTPFile{
		Path:     path,
		Requests: requests,
	}, nil
}

// ParseDirectory walks a directory and parses all .http files.
func (p *HTTPFileParser) ParseDirectory(dir string) ([]model.HTTPFile, error) {
	var files []model.HTTPFile
	err := p.fs.Walk(dir, func(path string, info ifs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walking %s: %w", path, err)
		}
		if info.IsDir() || !strings.HasSuffix(path, ".http") {
			return nil
		}
		httpFile, err := p.ParseFile(path)
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return fmt.Errorf("computing relative path for %s: %w", path, err)
		}
		httpFile.Path = relPath
		files = append(files, *httpFile)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// httpMethods contains the standard HTTP methods for detecting request lines.
var httpMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "PATCH": true,
	"DELETE": true, "HEAD": true, "OPTIONS": true, "TRACE": true, "CONNECT": true,
}

// isRequestLine reports whether a trimmed line starts with an HTTP method.
func isRequestLine(trimmed string) bool {
	return httpMethods[strings.SplitN(trimmed, " ", 2)[0]]
}

// ParseHTTPContent parses the text content of an .http file into requests.
// It scans lines looking for HTTP verb lines as request boundaries. Everything
// before a verb line is treated as comments/name; everything after is headers+body.
func ParseHTTPContent(content string) ([]model.HTTPRequest, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, model.ErrInvalidHTTPFormat
	}

	lines := strings.Split(content, "\n")
	var requests []model.HTTPRequest
	var commentLines []string

	for i := 0; i < len(lines); {
		trimmed := strings.TrimSpace(lines[i])

		// Skip blank lines between requests
		if trimmed == "" {
			i++
			continue
		}

		// Comment or separator line: accumulate for the next request name
		if strings.HasPrefix(trimmed, "#") {
			commentLines = append(commentLines, trimmed)
			i++
			continue
		}

		// Check for HTTP verb → start of a request
		if !isRequestLine(trimmed) {
			return nil, fmt.Errorf("%w: unexpected line: %q", model.ErrInvalidHTTPFormat, trimmed)
		}

		req, consumed, err := parseRequest(commentLines, lines[i:])
		if err != nil {
			return nil, err
		}
		requests = append(requests, *req)
		commentLines = nil
		i += consumed
	}

	if len(requests) == 0 {
		return nil, model.ErrInvalidHTTPFormat
	}
	return requests, nil
}

// parseRequest builds an HTTPRequest from accumulated comment lines and
// the remaining lines starting at the request (verb) line. Returns the
// request and how many lines from requestLines were consumed.
func parseRequest(comments []string, requestLines []string) (*model.HTTPRequest, int, error) {
	// Extract name from first comment
	name := extractName(comments)

	method, url, err := parseRequestLine(strings.TrimSpace(requestLines[0]))
	if err != nil {
		return nil, 0, err
	}

	if name == "" {
		name = method + " " + url
	}

	// Collect header and body lines until next request or end
	remaining := requestLines[1:]
	consumed := findNextRequest(remaining)
	headers, body := parseHeadersAndBody(remaining[:consumed])

	if strings.HasPrefix(url, "/") {
		for _, h := range headers {
			if strings.EqualFold(h.Key, "Host") {
				url = h.Value + url
				break
			}
		}
	}

	return &model.HTTPRequest{
		Name:    name,
		Method:  method,
		URL:     url,
		Headers: filterHostHeader(headers),
		Body:    body,
	}, consumed + 1, nil
}

// extractName returns the cleaned name from the first comment line.
func extractName(comments []string) string {
	for _, c := range comments {
		cleaned := cleanCommentName(c)
		if cleaned != "" {
			return cleaned
		}
	}
	return ""
}

// findNextRequest returns the number of lines to consume before the
// next request starts. When it hits a comment/separator line, it scans
// ahead for a following HTTP verb; if found, the comment belongs to
// the next request so we stop here.
func findNextRequest(lines []string) int {
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if isRequestLine(trimmed) {
			return i
		}
		if !strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Comment line: check if a verb follows (skipping blanks/comments).
		if nextVerbFollows(lines[i+1:]) {
			return i
		}
	}
	return len(lines)
}

// nextVerbFollows reports whether the next non-blank, non-comment line
// is an HTTP request line.
func nextVerbFollows(lines []string) bool {
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		return isRequestLine(trimmed)
	}
	return false
}

// parseRequestLine parses "METHOD URL" or "METHOD URL HTTP/1.1" into method and URL.
func parseRequestLine(line string) (string, string, error) {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return "", "", fmt.Errorf("%w: %q", model.ErrInvalidURLFormat, line)
	}
	method := parts[0]
	if !httpMethods[method] {
		return "", "", fmt.Errorf("%w: unknown method %q in %q", model.ErrInvalidURLFormat, method, line)
	}
	url := parts[1]
	// parts[2] would be "HTTP/1.1" if present — we ignore it
	return method, url, nil
}

// filterHostHeader removes Host headers since the URL already contains the host.
func filterHostHeader(headers []model.HTTPHeader) []model.HTTPHeader {
	var filtered []model.HTTPHeader
	for _, h := range headers {
		if !strings.EqualFold(h.Key, "Host") {
			filtered = append(filtered, h)
		}
	}
	return filtered
}

// cleanCommentName extracts a clean name from a comment line, stripping
// the # prefix and decorative characters like ─, ═, -, =, ~.
func cleanCommentName(line string) string {
	// Remove leading # characters and spaces
	name := strings.TrimLeft(line, "# ")
	// Remove decorative box-drawing and separator characters from both ends
	name = strings.Trim(name, "─═—–-=~_ ")
	// Clean up any remaining internal sequences of decorative chars
	// (e.g., "── Create Token ──" → "Create Token")
	name = strings.TrimSpace(name)
	return name
}

func parseHeadersAndBody(lines []string) ([]model.HTTPHeader, string) {
	var headers []model.HTTPHeader
	var bodyLines []string
	headersEnded := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if !headersEnded {
				headersEnded = true
			}
			continue
		}
		if headersEnded {
			bodyLines = append(bodyLines, line)
		} else {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				headers = append(headers, model.HTTPHeader{
					Key:   strings.TrimSpace(parts[0]),
					Value: strings.TrimSpace(parts[1]),
				})
			}
		}
	}

	return headers, strings.Join(bodyLines, "\n")
}
