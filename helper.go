package rexon

import (
	"context"
	"regexp"
)

var (
	rexParseSize        = RexCompile(`([-+]?[0-9]*\.?[0-9]+)\s*(\w+)?`)
	rexRemoveEmptyLines = RexCompile(`(?m)^\s*$[\r\n]*|[\r\n]+\s+\z`)
	emptyByte           = []byte("")
)

// RexSetCompile compiles a RexSet map
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
