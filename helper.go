package rexon

import (
	"context"
	"regexp"
)

var (
	rexParseSize        = regexp.MustCompile(`([-+]?[0-9]*\.?[0-9]+)\s*(\w+)?`)
	rexRemoveEmptyLines = regexp.MustCompile(`(?m)^\s*$[\r\n]*|[\r\n]+\s+\z`)
	emptyByte           = []byte("")
)

// RexCompile wraps regexp.Compile
func RexCompile(rex string) (*regexp.Regexp, error) {
	return regexp.Compile(rex)
}

// RexMustCompile wraps regexp.Compile
func RexMustCompile(rex string) *regexp.Regexp {
	return regexp.MustCompile(rex)
}

// RexSetCompile compiles a RexSet map
func RexSetCompile(set map[string]string) (map[string]*regexp.Regexp, error) {
	rexSet := make(map[string]*regexp.Regexp)
	for key, value := range set {
		rex, err := regexp.Compile(value)
		if err != nil {
			return nil, err
		}
		rexSet[key] = rex
	}
	return rexSet, nil
}

// RexSetMustCompile is like RexSetCompile but panics on error
func RexSetMustCompile(set map[string]string) map[string]*regexp.Regexp {
	rexSet, err := RexSetCompile(set)
	if err != nil {
		panic(err)
	}
	return rexSet
}

// wrapCtxSend wraps the sending to a channel with a context
func wrapCtxSend(ctx context.Context, result *Result, resultCh chan<- *Result) bool {
	select {
	case <-ctx.Done():
		return false
	case resultCh <- result:
		return true
	}
}

func getFieldType(field string, fieldTypes map[string]ValueType) (ValueType, bool) {

	if fieldTypes == nil {
		return "", false
	}

	// Return the specified ValueType
	if valueType, exists := fieldTypes[field]; exists {
		return valueType, exists
	}

	// Without a specific key, fallback to the catch all type if available
	if valueType, exists := fieldTypes[KeyTypeAll]; exists {
		return valueType, exists
	}

	return "", false
}
