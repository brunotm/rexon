package rexon

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
)

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
		var match [][]byte
		result := &Result{}

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
			parseField(result, p.Fields[i], p.FieldTypes, match[i], p.Round)
		}

		if result.Data != nil || result.Errors != nil {
			if !wrapCtxSend(ctx, result, resultCh) {
				return
			}
		}

	}
}
