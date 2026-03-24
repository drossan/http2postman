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

func parseHTTPSection(section string) (*model.HTTPRequest, error) {
	lines := strings.Split(section, "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("%w: section has fewer than 2 lines", model.ErrInvalidHTTPFormat)
	}

	name := strings.TrimPrefix(strings.TrimSpace(lines[0]), "# ")

	urlLine := strings.SplitN(strings.TrimSpace(lines[1]), " ", 2)
	if len(urlLine) < 2 {
		return nil, fmt.Errorf("%w: %q", model.ErrInvalidURLFormat, lines[1])
	}

	headers, body := parseHeadersAndBody(lines[2:])

	return &model.HTTPRequest{
		Name:    name,
		Method:  urlLine[0],
		URL:     urlLine[1],
		Headers: headers,
		Body:    body,
	}, nil
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
