package rexon

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Parse parses raw data using the specified RexSet
func (p *RexSet) Parse(ctx context.Context, data io.Reader) <-chan []byte {
	rexChan := make(chan []byte)
	go p.parse(ctx, data, rexChan)
	return rexChan
}

// ParseBytes parses raw data using the specified RexSetLine
func (p *RexSet) ParseBytes(ctx context.Context, data []byte) <-chan []byte {
	return p.Parse(ctx, bytes.NewReader(data))
}

func (p *RexSet) parse(ctx context.Context, data io.Reader, rexChan chan<- []byte) {
	defer close(rexChan)
	scanner := bufio.NewScanner(data)
	startTag := p.RexMap[KeyStartTag]
	dropTag := p.RexMap[KeyDropTag]
	skipTag := p.RexMap[KeySkipTag]
	continueTag := p.RexMap[KeyContinueTag]

	var ok bool
	var skip bool
	var valid bool
	var document []byte
	var valueType ValueType

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
		// deliver the document
		if startTag.Match(line) && valid {
			if !wrapCtxSend(ctx, document, rexChan) {
				return
			}
			document = nil
			valid = false
		}

		for key := range p.RexMap {

			// Skip control regexps
			if key == KeyStartTag || key == KeyDropTag || key == KeyContinueTag || key == KeySkipTag {
				continue
			}

			// Continue if we already have a match for this regexp
			// it should not happen. it is necessary?
			if gjson.GetBytes(document, key).Exists() {
				continue
			}

			// Continue if we don't match this regexp
			match := p.RexMap[key].FindSubmatch(line)
			if match == nil {
				continue
			}

			if len(match) > 2 {
				// Store the result as a [] under the given key if we have multiple matches
				// For now we don't support type parsing under [] objects, so it will be a []string
				document, _ = sjson.SetBytes(document, key, match[1:])
				if !valid {
					valid = true
				}
				continue
			}

			// // Only join match groups if keepmv is specified
			// if p.KeepMV {
			// 	document, _ = sjson.SetBytes(document, key, bytes.Join(match[1:], []byte(p.KeepMVSep)))
			// 	if !valid {
			// 		valid = true
			// 	}
			// }

			// If we have a type map prepare for parsing types
			if p.FieldTypes != nil {
				// Try to get a specific type for this key
				valueType, ok = p.FieldTypes[key]
				if !ok {
					// Without a specific key, fallback to the catch _all type, if available
					if valueType, ok = p.FieldTypes[KeyTypeAll]; !ok {
						// Else just set the resulting string and continue for the next key
						document, _ = sjson.SetBytes(document, key, match[1])
						if !valid {
							valid = true
						}
						continue
					}
				}
			} else {
				// If no field types specified, just set the resulting match
				document, _ = sjson.SetBytes(document, key, match[1])
				if !valid {
					valid = true
				}
				continue
			}

			// Try to parse the match to the specified ValueType
			if v, err := ParseValue(match[1], valueType, p.Round); err != nil {
				edocument, _ := sjson.SetBytes([]byte(""), KeyErrorMessage,
					fmt.Sprintf(parseErrorMessage, key, err.Error()))
				if !wrapCtxSend(ctx, edocument, rexChan) {
					return
				}
			} else {
				document, _ = sjson.SetBytes(document, key, v)
				if !valid {
					valid = true
				}
			}
		}
	}
	if valid {
		if !wrapCtxSend(ctx, document, rexChan) {
			return
		}
	}
}
