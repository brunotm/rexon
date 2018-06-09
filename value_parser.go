package rexon

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

// make sure we satisfy the rexon.ParserInterface
var _ ValueParser = (*Value)(nil)

type ValueType string

const (
	Number      ValueType = "number"
	String      ValueType = "string"
	Bool        ValueType = "bool"
	Time        ValueType = "time"
	Duration    ValueType = "duration"
	DigitalUnit ValueType = "digital_unit"

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
	rexUnit      = regexp.MustCompile(`([-+]?\d*\.?\d+)\s*([a-z,A-Z])?`)
	rexIsUnit    = regexp.MustCompile(`[-+]?\d*\.?\d+\s*[a-z,A-Z]`)
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

// Value represent each singular value to extract, parse and transform
type Value struct {
	name       string         // Value name
	nullable   bool           // Nullable
	valueType  ValueType      // ValueType
	fromFormat string         // Format to convert from
	toFormat   string         // Format to convert to
	round      int            // Round when parsing numbers
	regex      *regexp.Regexp // Regexp used to extract data
}

// ValueOpt functional options for Value
type ValueOpt func(*Value) (err error)

// NewValue creates a new value parser
func NewValue(name string, vt ValueType, options ...ValueOpt) (v *Value, err error) {
	v = &Value{name: name, valueType: vt, round: 2, nullable: false}
	for _, opt := range options {
		if err = opt(v); err != nil {
			return nil, err
		}
	}
	return v, nil
}

// MustNewValue is like NewValue but panics on error
func MustNewValue(name string, vt ValueType, options ...ValueOpt) (v *Value) {
	v, err := NewValue(name, vt, options...)
	if err != nil {
		panic(err)
	}
	return v
}

// ValueRegex sets the regexp for this value parser
func ValueRegex(expr string) (opt ValueOpt) {
	return func(v *Value) (err error) {
		r, err := regexp.Compile(expr)
		v.regex = r
		return err
	}
}

// Round sets the round for this value parser, defaults to 2 if not specified
func Round(round int) (opt ValueOpt) {
	return func(v *Value) (err error) {
		v.round = round
		return nil
	}
}

// FromFormat sets the from format for this value parser
func FromFormat(format string) (opt ValueOpt) {
	return func(v *Value) (err error) {
		v.fromFormat = format
		return nil
	}
}

// ToFormat sets the destination format for this value parser
func ToFormat(format string) (opt ValueOpt) {
	return func(v *Value) (err error) {
		v.toFormat = strings.ToLower(format)
		return nil
	}
}

// Nullable sets the value to null on parsing errors and ignores the error
func Nullable() (opt ValueOpt) {
	return func(v *Value) (err error) {
		v.nullable = true
		return nil
	}
}

// Name returns this value name
func (v *Value) Name() (name string) {
	return v.name
}

// Parse parses the value for the given byte. Uses the ValueRegex if specified to first extract the data
func (v *Value) Parse(b []byte) (value interface{}, matched bool, err error) {

	// Extract data with the provided regex if specified
	if v.regex != nil {
		match := v.regex.FindSubmatch(b)
		if match == nil {
			return nil, false, nil
		}

		if len(match) > 2 {
			return nil, true, fmt.Errorf("parser: %s invalid number of matches for: %s", v.name, string(b))
		}

		b = match[1]
	}

	switch v.valueType {
	case String:
		value, err = v.parseString(b)
	case Number:
		value, err = v.parseNumber(b)
	case Bool:
		value, err = strconv.ParseBool(*(*string)(unsafe.Pointer(&b)))
	case Time:
		value, err = v.parseTime(b)
	case Duration:
		value, err = v.parseDuration(b)
	case DigitalUnit:
		value, err = v.parseUnit(b)
	default:
		err = fmt.Errorf("unsupported type %s for: %s", v.valueType, v.name)
	}

	// Set to null if we cannot parse and Nullable is specified
	if err != nil && v.nullable {
		return nil, true, nil
	}

	return value, true, err
}

func (v *Value) parseString(b []byte) (value interface{}, err error) {
	str := *(*string)(unsafe.Pointer(&b))
	if v.nullable && (str == "null" || str == "") {
		value = nil
	} else {
		value = str
	}

	return value, err
}

// parseDuration parses a string representation of duration into a specified time unit or in a time.Duration
func (v *Value) parseDuration(b []byte) (value interface{}, err error) {
	b = bytes.ToLower(b)

	if !rexIsUnit.Match(b) {
		// s := *(*string)(unsafe.Pointer(&b))
		// if !unicode.IsLetter(rune(s[len(s)-1])) {
		// Defaults to seconds if no unit available
		if v.fromFormat == "" {
			b = append(b, 's')
		} else {
			b = append(b, v.fromFormat...)
		}
	}

	s := *(*string)(unsafe.Pointer(&b))
	d, err := time.ParseDuration(s)
	if err != nil {
		return nil, err
	}

	switch v.toFormat {
	case "nanoseconds", "nanosecond", "nano", "ns":
		value = d.Nanoseconds()
	case "milliseconds", "millisecond", "milli", "ms":
		value = d.Nanoseconds() / int64(time.Millisecond)
	case "seconds", "second", "sec", "s":
		value = d.Seconds()
	case "minutes", "minute", "min", "m":
		value = d.Minutes()
	case "hours", "hour", "h":
		value = d.Hours()
	case "string", "":
		value = d.String()
	default:
		err = fmt.Errorf("unsupported destination format for %s: %s", v.name, v.toFormat)
	}

	return value, err
}

// parseTime parses a string representation of time from the specified format into a specified format or in a time.Time
func (v *Value) parseTime(b []byte) (value interface{}, err error) {
	s := *(*string)(unsafe.Pointer(&b))
	t, err := time.Parse(v.fromFormat, s)
	if err != nil {
		return nil, err
	}

	switch v.toFormat {
	case "unix":
		value = t.Unix()
	case "unix_milli":
		value = t.UnixNano() / int64(time.Millisecond)
	case "unix_nano":
		value = t.UnixNano()
	case "rfc3339":
		value = t.Format(time.RFC3339)
	case "rfc3339nano", "string", "":
		value = t.Format(time.RFC3339Nano)
	default:
		err = fmt.Errorf("unsupported destination format for %s: %s", v.name, v.toFormat)

	}

	return value, err
}

// parseNumber parses a number string representation into a float64
func (v *Value) parseNumber(b []byte) (value interface{}, err error) {
	f, err := strconv.ParseFloat(*(*string)(unsafe.Pointer(&b)), 64)
	if err != nil {
		return nil, err
	}

	if v.round < 1 {
		return value, nil
	}

	return round(f, v.round), nil
}

// parseUnit parses a digital unit string representation into a float64 in
// bytes or any other unit format
func (v *Value) parseUnit(b []byte) (value float64, err error) {

	b = bytes.ToLower(b)
	match := rexUnit.FindSubmatch(b)
	if match == nil {
		return 0, fmt.Errorf("no digital unit match for %s: %s", v.name, string(b))
	}

	val, err := strconv.ParseFloat(*(*string)(unsafe.Pointer(&match[1])), 64)
	if err != nil {
		return 0, err
	}

	// Use fromFormat if specified and no unit is found
	u := *(*string)(unsafe.Pointer(&match[2]))
	if u == "" && v.fromFormat != "" {
		u = v.fromFormat
	}
	unit, ok := digitalUnits[u]
	if !ok {
		return 0, fmt.Errorf("cannot parse unit for %s: %s", v.name, u)
	}
	val = val * unit

	// Convert to the specified unit
	unit, ok = digitalUnits[v.toFormat]
	if !ok {
		return 0, fmt.Errorf("unsupported unit for %s: %s", v.name, v.toFormat)
	}

	return round(float64(val/unit), v.round), nil
}

// Round a float to the specified precision
func round(f float64, round int) (n float64) {
	shift := math.Pow(10, float64(round))
	return math.Floor((f*shift)+.5) / shift
}
