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
	Round   int                  // The floating point precision to parse numbers
	FindAll bool                 // Find all ocurrences in line
	Prep    *regexp.Regexp       // The regexp used to strip unwanted stuff (eg. `[(),;!"']`) before matching
	Rex     *regexp.Regexp       // The Regexp for match
	Fields  []string             // The ordered field names
	Types   map[string]ValueType // The Fields:Type for conversion
}

// NewLineParser creates a new set parser with the given configuration
func NewLineParser() *RexLine {
	return &RexLine{}
}

// SetRex to the current parser
func (p *RexLine) SetRex(rex string, findAll bool) {
	p.Prep = regexp.MustCompile(rex)
	p.FindAll = findAll
}

// SetKey and type of match
// this must be called sequentially and in order for the results of the match
func (p *RexLine) SetKey(key string, valueType ValueType) {
	p.Fields = append(p.Fields, key)
	p.Types[key] = valueType
}

// SetPrep to the current parser
func (p *RexLine) SetPrep(rex string) {
	p.Prep = regexp.MustCompile(rex)
}

// SetRound default rounding used for number parsing
func (p *RexLine) SetRound(round int) {
	p.Round = round
}

// Parse parses raw data from a io.Reader using the specified RexSetLine
func (p *RexLine) Parse(ctx context.Context, data io.Reader) (results <-chan Result) {
	resultCh := make(chan Result)
	go p.parse(ctx, data, resultCh)
	return resultCh
}

// ParseBytes parses raw data from a []byte using the specified RexSetLine
func (p *RexLine) ParseBytes(ctx context.Context, data []byte) (results <-chan Result) {
	return p.Parse(ctx, bytes.NewReader(data))
}

func (p *RexLine) parse(ctx context.Context, data io.Reader, results chan<- Result) {
	defer close(results)
	scanner := bufio.NewScanner(data)

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			result := Result{}
			result.Errors = append(result.Errors, err)
			wrapCtxSend(ctx, result, results)
			return
		}

		var match [][]byte
		result := Result{}

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

			if !wrapCtxSend(ctx, result, results) {
				return
			}
			continue
		}

		for i := range match {
			// Set and parse fields
			result = parseFieldValue(result, p.Fields[i], p.Types, match[i], p.Round)
		}

		if result.Data != nil || result.Errors != nil {
			if !wrapCtxSend(ctx, result, results) {
				return
			}
		}

	}
}
