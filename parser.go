package rexon

import (
	"context"
	"io"
)

// Unit for unit parsing
type Unit float64

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
