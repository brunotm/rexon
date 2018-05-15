package rexon

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var (
	rexDefaultStartTag      = regexp.MustCompile(`.`)
	errInvalidParsersNumber = fmt.Errorf("invalid number of matches and parsers")
)

// Rex type
type Rex struct {
	multiLine    bool
	TrimSpaces   bool
	StartTag     *regexp.Regexp
	StopTag      *regexp.Regexp
	SkipTag      *regexp.Regexp
	ContinueTag  *regexp.Regexp
	Regexp       *regexp.Regexp
	ValueParsers []*ValueParser `json:"value_parsers"`
}

// UnmarshalJSON creates a new parser from the JSON encoded configuration
func (p *Rex) UnmarshalJSON(data []byte) (err error) {
	type alias Rex
	aux := &struct {
		*alias
		StartTag    string `json:"start_tag"`
		StopTag     string `json:"stop_tag"`
		SkipTag     string `json:"skip_tag"`
		ContinueTag string `json:"continue_tag"`
		Regexp      string `json:"regexp"`
	}{
		alias: (*alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.StartTag != "" {
		p.StartTag, err = regexp.Compile(aux.StartTag)
		if err != nil {
			return err
		}
	} else {
		p.StartTag = rexDefaultStartTag
	}

	if aux.StopTag != "" {
		p.StopTag, err = regexp.Compile(aux.StopTag)
		if err != nil {
			return err
		}

	}
	if aux.SkipTag != "" {
		p.SkipTag, err = regexp.Compile(aux.SkipTag)
		if err != nil {
			return err
		}
	}
	if aux.ContinueTag != "" {
		p.ContinueTag, err = regexp.Compile(aux.ContinueTag)
		if err != nil {
			return err
		}
	}

	if aux.Regexp != "" {
		p.Regexp, err = regexp.Compile(aux.Regexp)
		if err != nil {
			return err
		}
	}

	if strings.HasPrefix(aux.Regexp, "(?m)") {
		p.multiLine = true
	}

	return nil
}

// SetStartTag to the current parser
func (p *Rex) SetStartTag(rex string) {
	p.StartTag = regexp.MustCompile(rex)
}

// SetStopTag to the current parser
func (p *Rex) SetStopTag(rex string) {
	p.StopTag = regexp.MustCompile(rex)
}

// SetSkipTag to the current parser
func (p *Rex) SetSkipTag(rex string) {
	p.SkipTag = regexp.MustCompile(rex)
}

// SetContinueTag to the current parser
func (p *Rex) SetContinueTag(rex string) {
	p.ContinueTag = regexp.MustCompile(rex)
}

// Parse parses raw data using the specified Rex
func (p *Rex) Parse(ctx context.Context, data io.Reader) (results <-chan Result) {
	resultCh := make(chan Result)

	if p.Regexp == nil {
		go p.parseSet(ctx, data, resultCh)
		return resultCh
	}

	go p.parse(ctx, data, resultCh)
	return resultCh
}

// ParseBytes parses raw data using the specified RexLine
func (p *Rex) ParseBytes(ctx context.Context, data []byte) (results <-chan Result) {
	return p.Parse(ctx, bytes.NewReader(data))
}

func (p *Rex) parse(ctx context.Context, data io.Reader, results chan<- Result) {
	defer close(results)

	var skip bool
	var line []byte
	var match [][]byte
	var result Result
	var buff bytes.Buffer
	scanner := bufio.NewScanner(data)

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			result = Result{}
			result.Errors = append(result.Errors, err)
			wrapCtxSend(ctx, result, results)
			return
		}

		if p.TrimSpaces {
			line = bytes.TrimSpace(scanner.Bytes())
		} else {
			line = scanner.Bytes()
		}

		if p.StopTag != nil && p.StopTag.Match(line) {
			break
		}

		// Only skip sections if both skip and continue regexps are set
		if p.SkipTag != nil && p.ContinueTag != nil {
			// Set the skip flag if KeySkipTag is set and match the current line
			if p.SkipTag.Match(line) {
				skip = true
			}
			//  Set the skip flag if KeyContinueTag is set and match the current line
			if p.ContinueTag.Match(line) {
				skip = false
			}

			if skip {
				continue
			}
		}

		// Buffer lines till match when multiline (?m)
		if p.multiLine {
			if buff.Len() > 0 {
				buff.WriteByte('\n')
			}
			buff.Write(line)
			match = p.Regexp.FindSubmatch(buff.Bytes())
		} else {
			match = p.Regexp.FindSubmatch(line)
		}

		if match == nil {
			continue
		}

		match = match[1:]
		if len(match) != len(p.ValueParsers) {
			result = Result{}
			result.Errors = append(result.Errors, errInvalidParsersNumber)
			wrapCtxSend(ctx, result, results)
			return
		}

		result = Result{}
		result.Data = newJSON()

		for vp := range p.ValueParsers {

			value, err := p.ValueParsers[vp].ParseType(match[vp])
			if err != nil {
				err = fmt.Errorf("error parsing %s, %s", p.ValueParsers[vp].Name, err.Error())
				result.Errors = append(result.Errors, err)
			}

			result.Data, _ = jsonSet(result.Data, value, p.ValueParsers[vp].Name)
		}

		if p.multiLine {
			buff.Reset()
		}

		if !wrapCtxSend(ctx, result, results) {
			return
		}
	}

}

func (p *Rex) parseSet(ctx context.Context, data io.Reader, results chan<- Result) {
	defer close(results)

	var skip bool
	var result Result
	scanner := bufio.NewScanner(data)

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			result = Result{}
			result.Errors = append(result.Errors, err)
			wrapCtxSend(ctx, result, results)
			return
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// Trim and prepare if set a prepare regexp
		if p.TrimSpaces {
			line = bytes.TrimSpace(line)
		}

		// Drop if KeyDropTag is set and match the current line
		if p.StopTag != nil && p.StopTag.Match(line) {
			break
		}

		// Only skip sections if both skip and continue regexps are set
		if p.SkipTag != nil && p.ContinueTag != nil {
			// Set the skip flag if KeySkipTag is set and match the current line
			if p.SkipTag.Match(line) {
				skip = true
			}
			//  Set the skip flag if KeyContinueTag is set and match the current line
			if p.ContinueTag.Match(line) {
				skip = false
			}

			if skip {
				continue
			}
		}

		// If content is a match for start_tag and
		// document is valid deliver the result
		if p.StartTag.Match(line) {
			if len(result.Data) > 0 || result.Errors != nil {
				if !wrapCtxSend(ctx, result, results) {
					return
				}
			}
			result = Result{}
			result.Data = newJSON()
		}

		for vp := range p.ValueParsers {

			// Continue if we already have a match for this regexp
			if jsonHas(result.Data, p.ValueParsers[vp].Name) {
				continue
			}

			// Continue if we don't match this regexp
			value, ok, err := p.ValueParsers[vp].Parse(line)
			if err != nil {
				result.Errors = append(
					result.Errors,
					fmt.Errorf("error parsing %s, %s", p.ValueParsers[vp].Name, err.Error()),
				)
				continue
			}

			if ok {
				result.Data, _ = jsonSet(result.Data, value, p.ValueParsers[vp].Name)
			}
		}
	}

	if result.Data != nil || result.Errors != nil {
		if !wrapCtxSend(ctx, result, results) {
			return
		}
	}
}
