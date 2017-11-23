package rexon

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"regexp"
)

// RexSet type
type RexSet struct {
	Round int                       // The floating point precision to parse numbers
	Prep  *regexp.Regexp            // The regexp used to strip unwanted stuff (eg. `[(),;!"']`) before matching
	Set   map[string]*regexp.Regexp // The set of Field:Regexp
	Types map[string]ValueType      // The Field:Type map for conversions
}

// NewSetParser creates a new set parser with the given configuration
func NewSetParser() *RexSet {
	return &RexSet{
		Set:   make(map[string]*regexp.Regexp),
		Types: make(map[string]ValueType),
	}
}

// AddRex to the current parser
func (p *RexSet) AddRex(key, rex string, valueType ValueType) {
	p.Set[key] = regexp.MustCompile(rex)
	p.Types[key] = valueType
}

// SetPrep to the current parser
func (p *RexSet) SetPrep(rex string) {
	p.Prep = regexp.MustCompile(rex)
}

// SetStartTag to the current parser
func (p *RexSet) SetStartTag(rex string) {
	p.Set[KeyStartTag] = regexp.MustCompile(rex)
}

// SetDropTag to the current parser
func (p *RexSet) SetDropTag(rex string) {
	p.Set[KeyDropTag] = regexp.MustCompile(rex)
}

// SetSkipTag to the current parser
func (p *RexSet) SetSkipTag(rex string) {
	p.Set[KeySkipTag] = regexp.MustCompile(rex)
}

// SetContinueTag to the current parser
func (p *RexSet) SetContinueTag(rex string) {
	p.Set[KeyContinueTag] = regexp.MustCompile(rex)
}

// SetRound default rounding used for number parsing
func (p *RexSet) SetRound(round int) {
	p.Round = round
}

// Parse parses raw data using the specified RexSet
func (p *RexSet) Parse(ctx context.Context, data io.Reader) (results <-chan Result) {
	resultCh := make(chan Result)
	go p.parse(ctx, data, resultCh)
	return resultCh
}

// ParseBytes parses raw data using the specified RexSetLine
func (p *RexSet) ParseBytes(ctx context.Context, data []byte) (results <-chan Result) {
	return p.Parse(ctx, bytes.NewReader(data))
}

func (p *RexSet) parse(ctx context.Context, data io.Reader, results chan<- Result) {
	defer close(results)

	var skip bool
	result := Result{}
	scanner := bufio.NewScanner(data)
	startTag := p.Set[KeyStartTag]
	dropTag := p.Set[KeyDropTag]
	skipTag := p.Set[KeySkipTag]
	continueTag := p.Set[KeyContinueTag]

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			result := Result{}
			result.Errors = append(result.Errors, err)
			wrapCtxSend(ctx, result, results)
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
				if !wrapCtxSend(ctx, result, results) {
					return
				}
			}
			result = Result{}
		}

		for key := range p.Set {

			// Skip control regexps
			if key == KeyStartTag || key == KeyDropTag || key == KeyContinueTag || key == KeySkipTag {
				continue
			}

			// Continue if we already have a match for this regexp
			if JSONExists(result.Data, key) {
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
				result.Data, _ = JSONSet(result.Data, match[1:], key)
				continue
			}

			// Set and parse fields
			result = parseFieldValue(result, key, p.Types, match[1], p.Round)
		}
	}

	if result.Data != nil || result.Errors != nil {
		if !wrapCtxSend(ctx, result, results) {
			return
		}
	}
}
