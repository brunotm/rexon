package rexon

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/tidwall/sjson"
)

// Parse parses raw data from a io.Reader using the specified RexSetLine
func (p *RexLine) Parse(ctx context.Context, data io.Reader) <-chan []byte {
	rexChan := make(chan []byte)
	go p.parse(ctx, data, rexChan)
	return rexChan
}

// ParseBytes parses raw data from a []byte using the specified RexSetLine
func (p *RexLine) ParseBytes(ctx context.Context, data []byte) <-chan []byte {
	return p.Parse(ctx, bytes.NewReader(data))
}

func (p *RexLine) parse(ctx context.Context, data io.Reader, rexChan chan<- []byte) {
	defer close(rexChan)
	scanner := bufio.NewScanner(data)
	var ok bool
	var valueType ValueType

	for scanner.Scan() {
		var match [][]byte
		line := bytes.TrimSpace(scanner.Bytes())
		if p.RexPrep != nil {
			line = p.RexPrep.ReplaceAll(line, emptyByte)
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
			edocument, _ := sjson.SetBytes([]byte(""), KeyErrorMessage,
				fmt.Sprintf("rexon: match doesn't have the number of required fields: %s", bytes.Join(match, []byte(","))))
			if !wrapCtxSend(ctx, edocument, rexChan) {
				return
			}
			continue
		}

		var document []byte
		for i := range match {
			// If we have a type map prepare for parsing types
			if p.FieldTypes != nil {
				// Try to get a specific type for this key
				valueType, ok = p.FieldTypes[p.Fields[i]]
				if !ok {
					// Without a specific key, fallback to the catch _all type, if available
					if valueType, ok = p.FieldTypes[KeyTypeAll]; !ok {
						// Else just set the resulting string
						document, _ = sjson.SetBytes(document, p.Fields[i], match[i])
						continue
					}
				}
			} else {
				document, _ = sjson.SetBytes(document, p.Fields[i], match[i])
				continue
			}

			// Parse the field type
			if v, err := ParseValue(match[i], valueType, p.Round); err != nil {
				edocument, _ := sjson.SetBytes([]byte(""), KeyErrorMessage,
					fmt.Sprintf(parseErrorMessage, p.Fields[i], err.Error()))
				if !wrapCtxSend(ctx, edocument, rexChan) {
					return
				}
				// continue
			} else {
				document, _ = sjson.SetBytes(document, p.Fields[i], v)
			}
		}
		if document != nil {
			if !wrapCtxSend(ctx, document, rexChan) {
				return
			}
		}
	}

}
