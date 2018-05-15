package rexon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type ValueType string

const (
	Null     ValueType = "null"
	Number   ValueType = "number"
	String   ValueType = "string"
	Bool     ValueType = "bool"
	Time     ValueType = "time"
	Duration ValueType = "duration"
	Object   ValueType = "object"
	Array    ValueType = "array"
	Unknown  ValueType = "unknown"

	// Decimal
	Byte = 1
	KB   = Byte * 1000
	MB   = KB * 1000
	GB   = MB * 1000
	TB   = GB * 1000
	PB   = TB * 1000

	// Binary
	KiB = Byte * 1024
	MiB = KiB * 1024
	GiB = MiB * 1024
	TiB = GiB * 1024
	PiB = TiB * 1024
)

var (
	rexUnits     = regexp.MustCompile(`([-+]?[0-9]*\.?[0-9]+)\s*(\w+)?`)
	digitalUnits = map[string]float64{
		"":          Byte,
		"b":         Byte,
		"byte":      Byte,
		"k":         KB,
		"kb":        KB,
		"kilo":      KB,
		"kilobyte":  KB,
		"kilobytes": KB,
		"m":         MB,
		"mb":        MB,
		"mega":      MB,
		"megabyte":  MB,
		"megabytes": MB,
		"g":         GB,
		"gb":        GB,
		"giga":      GB,
		"gigabyte":  GB,
		"gigabytes": GB,
		"t":         TB,
		"tb":        TB,
		"tera":      TB,
		"terabyte":  TB,
		"terabytes": TB,
		"p":         PB,
		"pb":        PB,
		"peta":      PB,
		"petabyte":  PB,
		"petabytes": PB,
		"ki":        KiB,
		"kib":       KiB,
		"kibibyte":  KiB,
		"kibibytes": KiB,
		"mi":        MiB,
		"mib":       MiB,
		"mebibyte":  MiB,
		"mebibytes": MiB,
		"gi":        GiB,
		"gib":       GiB,
		"gibibyte":  GiB,
		"gibibytes": GiB,
		"ti":        TiB,
		"tib":       TiB,
		"tebibyte":  TiB,
		"tebibytes": TiB,
		"pi":        PiB,
		"pib":       PiB,
		"pebibyte":  PiB,
		"pebibytes": PiB,
	}
)

// ValueParser represent each singular value to extract, parse and transform
type ValueParser struct {
	Name       string         `json:"name"`                  // Value name
	Type       ValueType      `json:"type,omitempty"`        // ValueType
	FromFormat string         `json:"from_format,omitempty"` // Format to convert from
	ToFormat   string         `json:"to_format,omitempty"`   // Format to convert to
	Round      int            `json:"round,omitempty"`       // Round when parsing numbers
	Regexp     *regexp.Regexp `json:"regexp,omitempty"`      // Regexp used to extract data
}

// UnmarshalJSON sets *v to a copy of data
func (v *ValueParser) UnmarshalJSON(data []byte) (err error) {
	type alias ValueParser
	aux := &struct {
		*alias
		Regexp string `json:"regexp"`
	}{
		alias: (*alias)(v),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Regexp != "" {
		v.Regexp, err = regexp.Compile(aux.Regexp)
	}
	return err
}

// Parse extracts the value from the given []byte and its type using the configured parameters
func (v *ValueParser) Parse(b []byte) (value interface{}, ok bool, err error) {
	match := v.Regexp.FindSubmatch(b)
	if match == nil {
		return nil, false, nil
	}

	if len(match) > 2 {
		return nil, true, fmt.Errorf("parser: %s invalid number of matches for: %s", v.Name, string(b))
	}

	value, err = v.ParseType(match[1])
	return value, true, err
}

// ParseType for a given []byte using the configured parameters
func (v *ValueParser) ParseType(b []byte) (value interface{}, err error) {
	switch v.Type {
	case String:
		value = string(b)
	case Number:
		value, err = v.parseNumber(b)
	case Bool:
		value, err = strconv.ParseBool(*(*string)(unsafe.Pointer(&b)))
	case Time:
		value, err = v.parseTime(b)
	case Duration:
		value, err = v.parseDuration(b)
	default:
		err = fmt.Errorf("unsupported type %s for: %s", v.Type, v.Name)
	}

	return value, err
}

// parseDuration parses a string representation of duration into a specified time unit or in a time.Duration
func (v *ValueParser) parseDuration(b []byte) (value interface{}, err error) {
	s := *(*string)(unsafe.Pointer(&b))
	d, err := time.ParseDuration(s)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(v.ToFormat) {
	case "nanoseconds", "nanosecond", "nano":
		value = d.Nanoseconds()
	case "milliseconds", "millisecond", "milli":
		value = d.Nanoseconds() / int64(time.Millisecond)
	case "seconds", "second", "sec":
		value = d.Seconds()
	case "minutes", "minute", "min":
		value = d.Minutes()
	case "hours", "hour":
		value = d.Hours()
	case "string":
		value = d.String()
	default:
		err = fmt.Errorf("unsupported destination format for %s: %s", v.Name, v.ToFormat)
	}

	return value, err
}

// parseTime parses a string representation of time from the specified format into a specified format or in a time.Time
func (v *ValueParser) parseTime(b []byte) (value interface{}, err error) {
	s := *(*string)(unsafe.Pointer(&b))
	t, err := time.Parse(v.FromFormat, s)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(v.ToFormat) {
	case "unix":
		value = t.Unix()
	case "unix_milli":
		value = t.UnixNano() / int64(time.Millisecond)
	case "unix_nano":
		value = t.UnixNano()
	case "rfc3339nano":
		value = t.Format(time.RFC3339Nano)
	case "rfc3339", "string":
		value = t.Format(time.RFC3339)
	default:
		err = fmt.Errorf("unsupported destination format for %s: %s", v.Name, v.ToFormat)

	}

	return value, err
}

// parseNumber parses a number string representation into a float64
func (v *ValueParser) parseNumber(b []byte) (value float64, err error) {
	if v.FromFormat == "digital_unit" {
		return v.parseUnit(b)
	}

	return parseFloat64(*(*string)(unsafe.Pointer(&b)), v.Round)
}

// parseUnit parses a digital unit string representation into a float64 in
// bytes or any other unit format
func (v *ValueParser) parseUnit(b []byte) (value float64, err error) {

	b = bytes.ToLower(b)

	match := rexUnits.FindSubmatch(b)
	if match == nil {
		return 0, fmt.Errorf("no digital unit match for %s: %s", v.Name, string(b))
	}

	val, err := strconv.ParseFloat(*(*string)(unsafe.Pointer(&match[1])), 64)
	if err != nil {
		return 0, err
	}

	// Find the current unit and convert to bytes
	u := *(*string)(unsafe.Pointer(&match[2]))
	unit, ok := digitalUnits[u]
	if !ok {
		return 0, fmt.Errorf("cannot parse unit for %s: %s", v.Name, u)
	}
	val = val * unit

	// Convert to the specified unit
	unit, ok = digitalUnits[v.ToFormat]
	if !ok {
		return 0, fmt.Errorf("unsupported unit for %s: %s", v.Name, v.ToFormat)
	}
	return Round(float64(val/unit), v.Round), nil
}

// parseFloat64 parses a string into a float64 rounding it to the round precision
func parseFloat64(s string, r int) (f float64, err error) {
	f, err = strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}

	if r < 0 {
		return f, nil
	}
	return Round(f, r), nil
}

// Round a float to the specified precision
func Round(f float64, round int) (n float64) {
	shift := math.Pow(10, float64(round))
	return math.Floor((f*shift)+.5) / shift
}
