package rexon

import (
	"context"
	"io"
	"regexp"
)

const (
	// KeyStartTag identifies in RexSet.RexMap the regexp used to start matching for a new document
	KeyStartTag = "_start_tag"
	// KeyDropTag identifies in RexSet.RexMap the regexp used to stop parsing immediately
	KeyDropTag = "_drop_tag"
	// KeySkipTag identifies in RexSet.RexMap the regexp used to begin skipping data until KeyContinueTag
	KeySkipTag = "_skip_tag"
	// KeyContinueTag identifies in RexSet.RexMap the regexp used to continue matching after KeySkipTag
	KeyContinueTag = "_continue_tag"
	// KeyTypeAll identifies in RexLine/Set.FieldTypes the catch all type for type parsing (TypeInt, TypeFloat, TypeBool, TypeString)
	KeyTypeAll = "_all"
	// KeyErrorMessage identifies the parser error message in the resulting []byte json document
	KeyErrorMessage = "_error_message"
	// parseErrorMessage
	parseErrorMessage = "rexson: parsing %s error: %s"
)

// ValueType for type assertions
type ValueType uint16

// RexSON interface
type Parser interface {
	Parse(ctx context.Context, data io.Reader) <-chan []byte
	ParseBytes(ctx context.Context, data []byte) <-chan []byte
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

// RexSetCompile compiles a RexSet
func RexSetCompile(set map[string]string) map[string]*regexp.Regexp {
	rexSet := make(map[string]*regexp.Regexp)
	for key, value := range set {
		rex := regexp.MustCompile(value)
		rexSet[key] = rex
	}
	return rexSet
}

// RexCompile wraps regexp.MustCompile
func RexCompile(rex string) *regexp.Regexp {
	return regexp.MustCompile(rex)
}
