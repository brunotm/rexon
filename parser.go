package rexon

import (
	"context"
	"io"
)

// ValueType for type parsing
type ValueType string

// Unit for unit parsing
type Unit float64

// Result type from Parse methods
type Result struct {
	Errors []error
	Data   []byte
}

// Parser interface
type Parser interface {
	Parse(ctx context.Context, data io.Reader) <-chan *Result
	ParseBytes(ctx context.Context, data []byte) <-chan *Result
}
