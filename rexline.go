package rexon

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
)

// RexLine type
type RexLine struct {
	Round   int                  // The floating point precision to use when converting to float
	FindAll bool                 // Find all ocurrences in line
	Prep    *regexp.Regexp       // The regexp used to prepare (ReplaceAll) each line before matching (eg. `[(),!"]`)
	Rex     *regexp.Regexp       // The Regexp for match
	Fields  []string             // The ordered field names
	Types   map[string]ValueType // The Fields:Type for conversion
}

// NewLineParser creates a new set parser with the given configuration
func NewLineParser(prep, rex string, fields []string, findAll bool, types map[string]ValueType, round int) (Parser, error) {
	var err error

	if rex == "" {
		return nil, fmt.Errorf("rexson: empty rex")
	}

	// Check for empty or nil fields
	if len(fields) < 1 {
		return nil, fmt.Errorf("rexson: invalid fields %#v", fields)
	}

	parser := &RexLine{}
	parser.Round = round
	parser.Types = types
	parser.FindAll = findAll

	parser.Rex, err = RexCompile(rex)
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

// MustLineParser is like NewLineParser but panics on error
func MustLineParser(prep, rex string, fields []string, findAll bool, types map[string]ValueType, round int) Parser {
	parser, err := NewLineParser(prep, rex, fields, findAll, types, round)
	if err != nil {
		panic(err)
	}
	return parser
}

// Parse parses raw data from a io.Reader using the specified RexSetLine
func (p *RexLine) Parse(ctx context.Context, data io.Reader) <-chan *Result {
	resultCh := make(chan *Result)
	go p.parse(ctx, data, resultCh)
	return resultCh
}

// ParseBytes parses raw data from a []byte using the specified RexSetLine
func (p *RexLine) ParseBytes(ctx context.Context, data []byte) <-chan *Result {
	return p.Parse(ctx, bytes.NewReader(data))
}

func (p *RexLine) parse(ctx context.Context, data io.Reader, resultCh chan<- *Result) {
	defer close(resultCh)
	scanner := bufio.NewScanner(data)

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			result := &Result{}
			result.Errors = append(result.Errors, err)
			wrapCtxSend(ctx, result, resultCh)
			return
		}

		var match [][]byte
		result := &Result{}

		line := bytes.TrimSpace(scanner.Bytes())
		if p.Prep != nil {
			line = p.Prep.ReplaceAll(line, emptyByte)
		}

		if p.FindAll {
			// Search all ocurrences of the capturing group
			// for shorter repeating regexps
			subs := p.Rex.FindAllSubmatch(line, -1)
			if subs == nil || len(subs) == 0 {
				continue
			}
			for m := range subs {
				match = append(match, subs[m][1])
			}
		} else {
			// Search for non repeating capturing groups
			// for this case we match once the specified group in regexp
			// similiar with the set parser
			match = p.Rex.FindSubmatch(line)
			if match == nil {
				continue
			}
			// Drop the full match from the result
			match = match[1:]
		}

		if len(match) != len(p.Fields) {
			result.Errors = append(result.Errors,
				fmt.Errorf("rexon: match doesn't have the number of required fields: %s",
					bytes.Join(match, []byte(";"))))

			if !wrapCtxSend(ctx, result, resultCh) {
				return
			}
			continue
		}

		for i := range match {
			// Set and parse fields
			parseField(result, p.Fields[i], p.Types, match[i], p.Round)
		}

		if result.Data != nil || result.Errors != nil {
			if !wrapCtxSend(ctx, result, resultCh) {
				return
			}
		}

	}
}
