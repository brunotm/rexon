package rexon

import (
	"bufio"
	"bytes"
	"context"
	"io"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

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
	startTag := p.RexMap[KeyStartTag]
	dropTag := p.RexMap[KeyDropTag]
	skipTag := p.RexMap[KeySkipTag]
	continueTag := p.RexMap[KeyContinueTag]

	for scanner.Scan() {
		// Trim and prepare if set a prepare regexp
		line := bytes.TrimSpace(scanner.Bytes())
		if p.RexPrep != nil {
			line = p.RexPrep.ReplaceAll(line, emptyByte)
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

		for key := range p.RexMap {

			// Skip control regexps
			if key == KeyStartTag || key == KeyDropTag || key == KeyContinueTag || key == KeySkipTag {
				continue
			}

			// Continue if we already have a match for this regexp
			if gjson.GetBytes(result.Data, key).Exists() {
				continue
			}

			// Continue if we don't match this regexp
			match := p.RexMap[key].FindSubmatch(line)
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
			parseField(result, key, p.FieldTypes, match[1], p.Round)
		}
	}

	if result.Data != nil || result.Errors != nil {
		if !wrapCtxSend(ctx, result, resultCh) {
			return
		}
	}
}
