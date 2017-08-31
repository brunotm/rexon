package rexon

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// RexSet type
type RexSet struct {
	Round int                       // The floating point precision to use when converting to float
	Prep  *regexp.Regexp            // The regexp used to prepare (ReplaceAll) each line before matching (eg. `[(),;!"']`)
	Set   map[string]*regexp.Regexp // The set of Field:Regexp
	Types map[string]ValueType      // The Fields:Type for conversion
}

// NewSetParser creates a new set parser with the given configuration
func NewSetParser(prep string, set map[string]string, types map[string]ValueType, round int) (Parser, error) {
	var err error

	// Check for empty or nil set
	if len(set) < 1 {
		return nil, fmt.Errorf("rexson: invalid set %#v", set)
	}

	// Check if a start tag was provided
	if _, ok := set[KeyStartTag]; !ok {
		return nil, fmt.Errorf("rexson: set does not contain a start tag %#v", set)
	}

	parser := &RexSet{}
	parser.Round = round
	parser.Types = types

	parser.Set, err = RexSetCompile(set)
	if err != nil {
		return nil, err
	}

	// Check if we're given a prepare regexp and compile
	if prep != "" {
		parser.Prep, err = RexCompile(prep)
		if err != nil {
			return nil, err
		}
	}

	return parser, nil
}

// MustSetParser is like NewSetParser but panics on error
func MustSetParser(prep string, set map[string]string, types map[string]ValueType, round int) Parser {
	parser, err := NewSetParser(prep, set, types, round)
	if err != nil {
		panic(err)
	}
	return parser
}

// Parse parses raw data using the specified RexSet
func (p *RexSet) Parse(ctx context.Context, data io.Reader) <-chan *Result {
	resultCh := make(chan *Result)
	go p.parse(ctx, data, resultCh)
	return resultCh
}

// ParseBytes parses raw data using the specified RexSetLine
func (p *RexSet) ParseBytes(ctx context.Context, data []byte) <-chan *Result {
	return p.Parse(ctx, bytes.NewReader(data))
}

func (p *RexSet) parse(ctx context.Context, data io.Reader, resultCh chan<- *Result) {
	defer close(resultCh)

	var skip bool
	result := &Result{}
	scanner := bufio.NewScanner(data)
	startTag := p.Set[KeyStartTag]
	dropTag := p.Set[KeyDropTag]
	skipTag := p.Set[KeySkipTag]
	continueTag := p.Set[KeyContinueTag]

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			result := &Result{}
			result.Errors = append(result.Errors, err)
			wrapCtxSend(ctx, result, resultCh)
			return
		}

		// Trim and prepare if set a prepare regexp
		line := bytes.TrimSpace(scanner.Bytes())
		if p.Prep != nil {
			line = p.Prep.ReplaceAll(line, emptyByte)
		}

		// Drop if KeyDropTag is set and match the current line
		if dropTag != nil && dropTag.Match(line) {
			break
		}

		// Only skip sections if both skip and continue regexps are set
		if skipTag != nil && continueTag != nil {
			// Set the skip flag if KeySkipTag is set and match the current line
			if skipTag.Match(line) {
				skip = true
			}
			//  Set the skip flag if KeyContinueTag is set and match the current line
			if continueTag.Match(line) {
				skip = false
			}

			if skip {
				continue
			}
		}

		// If content is a match for start_tag and document is valid
		// deliver the result
		if startTag.Match(line) {
			if result.Data != nil || result.Errors != nil {
				if !wrapCtxSend(ctx, result, resultCh) {
					return
				}
			}
			result = &Result{}
		}

		for key := range p.Set {

			// Skip control regexps
			if key == KeyStartTag || key == KeyDropTag || key == KeyContinueTag || key == KeySkipTag {
				continue
			}

			// Continue if we already have a match for this regexp
			if gjson.GetBytes(result.Data, key).Exists() {
				continue
			}

			// Continue if we don't match this regexp
			match := p.Set[key].FindSubmatch(line)
			if match == nil {
				continue
			}

			if len(match) > 2 {
				// Store the result as a [] under the given key if we have multiple matches
				// TODO: add support for conversion under json arrays
				result.Data, _ = sjson.SetBytes(result.Data, key, match[1:])
				continue
			}

			// // Only join match groups if keepmv is specified
			// if p.KeepMV {
			// 	document, _ = sjson.SetBytes(document, key, bytes.Join(match[1:], []byte(p.KeepMVSep)))
			// continue
			// }

			// Set and parse fields
			parseField(result, key, p.Types, match[1], p.Round)
		}
	}

	if result.Data != nil || result.Errors != nil {
		if !wrapCtxSend(ctx, result, resultCh) {
			return
		}
	}
}
