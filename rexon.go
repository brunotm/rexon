package rexon

import (
	"context"
	"io"
)

// Result type from Parse methods
type Result struct {
	Data   []byte
	Errors []error
}

// Parser interface
type Parser interface {
	Parse(ctx context.Context, data io.Reader) <-chan Result
	ParseBytes(ctx context.Context, data []byte) <-chan Result
}
