package model

import (
	"errors"
	"fmt"
	"testing"
)

func TestSentinelErrors_ErrorsIs(t *testing.T) {
	tests := []struct {
		name     string
		sentinel error
	}{
		{"ErrInvalidHTTPFormat", ErrInvalidHTTPFormat},
		{"ErrInvalidURLFormat", ErrInvalidURLFormat},
		{"ErrInvalidCollection", ErrInvalidCollection},
		{"ErrEnvFileNotFound", ErrEnvFileNotFound},
		{"ErrEmptyCollectionName", ErrEmptyCollectionName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := fmt.Errorf("context: %w", tt.sentinel)
			if !errors.Is(wrapped, tt.sentinel) {
				t.Errorf("errors.Is failed for wrapped %s", tt.name)
			}
		})
	}
}
