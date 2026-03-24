package model

// HTTPFile represents a parsed .http file with its requests.
type HTTPFile struct {
	Path     string        // Relative path to the base directory
	Requests []HTTPRequest
}

// HTTPRequest represents a single HTTP request within a .http file.
type HTTPRequest struct {
	Name    string
	Method  string
	URL     string
	Headers []HTTPHeader
	Body    string
}

// HTTPHeader represents an HTTP header key-value pair.
type HTTPHeader struct {
	Key   string
	Value string
}
