package rexon

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var (
	rexDefaultStartTag      = regexp.MustCompile(`.`)
	errInvalidParsersNumber = fmt.Errorf("invalid number of matches and parsers")
)

// make sure we satisfy the rexon.ParserInterface
var _ DataParser = (*Parser)(nil)

// Parser type
type Parser struct {
	multiLine   bool
	trimSpaces  bool
	startTag    *regexp.Regexp
	stopTag     *regexp.Regexp
	skipTag     *regexp.Regexp
	continueTag *regexp.Regexp
	regex       *regexp.Regexp
	values      []*Value
}

// ParserOpt functional options for Parser
type ParserOpt func(*Parser) (err error)

// NewParser creates a new Parser
func NewParser(values []*Value, options ...ParserOpt) (p *Parser, err error) {
	p = &Parser{}
	p.startTag = rexDefaultStartTag

	for _, opt := range options {
		if err = opt(p); err != nil {
			return nil, err
		}
	}
	p.values = values
	return p, nil
}

// MustNewParser is like NewParser but panics on error
func MustNewParser(values []*Value, options ...ParserOpt) (p *Parser) {
	p, err := NewParser(values, options...)
	if err != nil {
		panic(err)
	}
	return p
}

// TrimSpaces trims right and left trailing spaces
func TrimSpaces() (opt ParserOpt) {
	return func(p *Parser) (err error) {
		p.trimSpaces = true
		return nil
	}
}

// LineRegex sets a regexp for this parser making the extraction to work in line mode and ignore all Values regexps.
// Working in line mode is much faster than in Set using Value regexp.
// Multiline regexps `(?m)` are still valid, but usually are slower than using Values regexps.
func LineRegex(expr string) (opt ParserOpt) {
	return func(p *Parser) (err error) {
		regex, err := regexp.Compile(expr)
		p.regex = regex
		return err
	}
}

// StartTag sets a regexp that will be used to start the match and extract
// when working in Set mode (Value regexp)
func StartTag(expr string) (opt ParserOpt) {
	return func(p *Parser) (err error) {
		regex, err := regexp.Compile(expr)
		p.startTag = regex
		return err
	}
}

// StopTag sets a regexp that when match will stop the parser
func StopTag(expr string) (opt ParserOpt) {
	return func(p *Parser) (err error) {
		regex, err := regexp.Compile(expr)
		p.stopTag = regex
		return err
	}
}

// SkipTag sets a regexp that when match the parser will skip lines until ContinueTag
func SkipTag(expr string) (opt ParserOpt) {
	return func(p *Parser) (err error) {
		regex, err := regexp.Compile(expr)
		p.skipTag = regex
		return err
	}
}

// ContinueTag sets a regexp that when match the parser will resume after SkipTag
func ContinueTag(expr string) (opt ParserOpt) {
	return func(p *Parser) (err error) {
		regex, err := regexp.Compile(expr)
		p.continueTag = regex
		return err
	}
}

// Parse parses raw data using the specified Rex
func (p *Parser) Parse(ctx context.Context, data io.Reader) (results <-chan Result) {
	resultCh := make(chan Result)

	if p.regex == nil {
		go p.parseSet(ctx, data, resultCh)
		return resultCh
	}

	go p.parse(ctx, data, resultCh)
	return resultCh
}

// ParseBytes parses raw data using the specified RexLine
func (p *Parser) ParseBytes(ctx context.Context, data []byte) (results <-chan Result) {
	return p.Parse(ctx, bytes.NewReader(data))
}

func (p *Parser) parse(ctx context.Context, data io.Reader, results chan<- Result) {
	defer close(results)

	var skip bool
	var line []byte
	var match [][]byte
	var result Result
	var buff bytes.Buffer
	scanner := bufio.NewScanner(data)

	// Handle multiline regexps
	if strings.HasPrefix(p.regex.String(), "(?m)") {
		p.multiLine = true
	}

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			result = Result{}
			result.Errors = append(result.Errors, err)
			wrapCtxSend(ctx, result, results)
			return
		}

		if p.trimSpaces {
			line = bytes.TrimSpace(scanner.Bytes())
		} else {
			line = scanner.Bytes()
		}

		if p.stopTag != nil && p.stopTag.Match(line) {
			break
		}

		// Only skip sections if both skip and continue regexps are set
		if p.skipTag != nil && p.continueTag != nil {
			// Set the skip flag if skipTag is set and match the current line
			if p.skipTag.Match(line) {
				skip = true
			}
			//  Set the skip flag if continueTag is set and match the current line
			if p.continueTag.Match(line) {
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
			match = p.regex.FindSubmatch(buff.Bytes())
		} else {
			match = p.regex.FindSubmatch(line)
		}

		if match == nil {
			continue
		}

		match = match[1:]
		if len(match) != len(p.values) {
			result = Result{}
			result.Errors = append(result.Errors, errInvalidParsersNumber)
			wrapCtxSend(ctx, result, results)
			return
		}

		result = Result{}
		result.Data = newJSON()

		for vp := range p.values {

			value, err := p.values[vp].ParseType(match[vp])
			if err != nil {
				err = fmt.Errorf("error parsing %s, %s", p.values[vp].name, err.Error())
				result.Errors = append(result.Errors, err)
			}

			result.Data, _ = jsonSet(result.Data, value, p.values[vp].name)
		}

		if p.multiLine {
			buff.Reset()
		}

		if !wrapCtxSend(ctx, result, results) {
			return
		}
	}

}

func (p *Parser) parseSet(ctx context.Context, data io.Reader, results chan<- Result) {
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

		if p.trimSpaces {
			line = bytes.TrimSpace(line)
		}

		if p.stopTag != nil && p.stopTag.Match(line) {
			break
		}

		// Only skip sections if both skip and continue regexps are set
		if p.skipTag != nil && p.continueTag != nil {
			// Set the skip flag if KeySkipTag is set and match the current line
			if p.skipTag.Match(line) {
				skip = true
			}
			//  Set the skip flag if KeyContinueTag is set and match the current line
			if p.continueTag.Match(line) {
				skip = false
			}

			if skip {
				continue
			}
		}

		// If content is a match for start_tag and
		// document is valid deliver the result
		if p.startTag.Match(line) {
			if len(result.Data) > 0 || result.Errors != nil {
				if !wrapCtxSend(ctx, result, results) {
					return
				}
			}
			result = Result{}
			result.Data = newJSON()
		}

		for vp := range p.values {

			// Continue if we already have a match for this regexp
			if jsonHas(result.Data, p.values[vp].name) {
				continue
			}

			// Continue if we don't match this regexp
			value, ok, err := p.values[vp].Parse(line)
			if err != nil {
				result.Errors = append(
					result.Errors,
					fmt.Errorf("error parsing %s, %s", p.values[vp].name, err.Error()),
				)
				continue
			}

			if ok {
				result.Data, _ = jsonSet(result.Data, value, p.values[vp].name)
			}
		}
	}

	if result.Data != nil || result.Errors != nil {
		if !wrapCtxSend(ctx, result, results) {
			return
		}
	}
}
