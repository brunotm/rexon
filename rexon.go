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

// DataParser interface
type DataParser interface {
	Parse(ctx context.Context, data io.Reader) (results <-chan Result)
	ParseBytes(ctx context.Context, data []byte) (results <-chan Result)
}

// ValueParser interface
type ValueParser interface {
	Parse(b []byte) (value interface{}, ok bool, err error)
	ParseType(b []byte) (value interface{}, err error)
}
