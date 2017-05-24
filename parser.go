package rexon

import (
	"context"
	"io"
	"regexp"
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

// RexSet type
type RexSet struct {
	Round      int                       // The floating point precision to use when converting to float
	RexPrep    *regexp.Regexp            // The regexp used to prepare (ReplaceAll) each line before matching (eg. `[(),;!"']`)
	RexMap     map[string]*regexp.Regexp // The set of Field:Regexp
	FieldTypes map[string]ValueType      // The Fields:Type for conversion
}

// RexLine type
type RexLine struct {
	Round      int                  // The floating point precision to use when converting to float
	RexPrep    *regexp.Regexp       // The regexp used to prepare (ReplaceAll) each line before matching (eg. `[(),!"]`)
	FindAll    bool                 // Find all ocurrences in line
	Rex        *regexp.Regexp       // The Regexp for match
	Fields     []string             // The ordered field names
	FieldTypes map[string]ValueType // The Fields:Type for conversion
}
