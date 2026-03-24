package parser

import (
	"fmt"
	ifs "io/fs"
	"path/filepath"
	"strings"

	"github.com/drossan/http2postman/internal/fs"
	"github.com/drossan/http2postman/internal/model"
)

const httpSectionSeparator = "###"

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

// ParseHTTPContent parses the text content of an .http file into requests.
func ParseHTTPContent(content string) ([]model.HTTPRequest, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, model.ErrInvalidHTTPFormat
	}

	sections := strings.Split(content, httpSectionSeparator)
	var requests []model.HTTPRequest

	for _, section := range sections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}
		req, err := parseHTTPSection(section)
		if err != nil {
			return nil, err
		}
		requests = append(requests, *req)
	}

	if len(requests) == 0 {
		return nil, model.ErrInvalidHTTPFormat
	}
	return requests, nil
}

// httpMethods contains the standard HTTP methods for detecting request lines.
var httpMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "PATCH": true,
	"DELETE": true, "HEAD": true, "OPTIONS": true, "TRACE": true, "CONNECT": true,
}

func parseHTTPSection(section string) (*model.HTTPRequest, error) {
	lines := strings.Split(section, "\n")
	if len(lines) < 1 {
		return nil, fmt.Errorf("%w: empty section", model.ErrInvalidHTTPFormat)
	}

	// Find the request line (METHOD URL), skipping comment lines (#) and blank lines.
	// The first comment line becomes the request name.
	var name string
	requestLineIdx := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		firstWord := strings.SplitN(trimmed, " ", 2)[0]
		if httpMethods[firstWord] {
			requestLineIdx = i
			break
		}
		if strings.HasPrefix(trimmed, "#") {
			// Use the first comment as name; skip subsequent comments
			if name == "" {
				name = strings.TrimPrefix(trimmed, "# ")
				name = strings.TrimPrefix(name, "#")
				name = strings.TrimSpace(name)
			}
			continue
		}
		return nil, fmt.Errorf("%w: unexpected line before request: %q", model.ErrInvalidHTTPFormat, trimmed)
	}

	if requestLineIdx < 0 {
		return nil, fmt.Errorf("%w: no request line found", model.ErrInvalidHTTPFormat)
	}

	method, url, err := parseRequestLine(strings.TrimSpace(lines[requestLineIdx]))
	if err != nil {
		return nil, err
	}

	// If no comment name, derive name from "METHOD URL"
	if name == "" {
		name = method + " " + url
	}

	// Check if Host header provides the base URL (for relative paths like "/api/users")
	remainingLines := lines[requestLineIdx+1:]
	headers, body := parseHeadersAndBody(remainingLines)

	// If URL is a relative path, prepend the Host header value
	if strings.HasPrefix(url, "/") {
		for _, h := range headers {
			if strings.EqualFold(h.Key, "Host") {
				url = h.Value + url
				break
			}
		}
	}

	// Filter out Host header from the exported headers (Postman uses URL, not Host)
	filteredHeaders := filterHostHeader(headers)

	return &model.HTTPRequest{
		Name:    name,
		Method:  method,
		URL:     url,
		Headers: filteredHeaders,
		Body:    body,
	}, nil
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
