package model

import "errors"

var (
	ErrInvalidHTTPFormat   = errors.New("invalid HTTP request format")
	ErrInvalidURLFormat    = errors.New("invalid URL line format")
	ErrInvalidCollection   = errors.New("invalid Postman collection format")
	ErrEnvFileNotFound     = errors.New("http-client.env.json not found")
	ErrEmptyCollectionName = errors.New("collection name cannot be empty")
)
