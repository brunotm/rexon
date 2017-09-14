package rexon

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"unsafe"
)

// ParseJsonValues parses values in a []byte encoded json document to the type specified type
// for each key in the fieldTypes argument. Ommited fields will resort to the KeyTypeAll type
// if specified, else the current value and type will remain unaltered.
func ParseJsonValues(data []byte, fieldTypes map[string]ValueType, round int) Result {
	result := Result{}

	// Iterate over document key and values for
	JSONForEach(data, func(key string, value []byte, tp ValueType) error {
		// Set and parse fields
		result = parseFieldValue(result, key, fieldTypes, value, round)
		return nil
	})

	return result
}

func parseFieldValue(result Result, field string, fieldTypes map[string]ValueType, match []byte, round int) Result {
	// If we have a type map prepare for parsing types
	if fieldType, ok := getFieldType(field, fieldTypes); ok {

		// Try to parse the match to the specified ValueType
		if value, err := parseValue(match, fieldType, round); err != nil {

			// Append parse error to result.Errors and set field to null
			result.Errors = append(
				result.Errors,
				fmt.Errorf("rexon: error parsing key %s to %s: %s", field, fieldType, err))

			result.Data, _ = JSONSet(result.Data, nil, field)

		} else {
			result.Data, _ = JSONSet(result.Data, value, field)
		}
	} else {

		// If no field types specified, just set the resulting match
		result.Data, _ = JSONSet(result.Data, match, field)
	}

	return result
}

// parseValue parses the given []byte to the specified ValueType.
// The round argument is only used when the ValueType is TypeFloat.
func parseValue(b []byte, t ValueType, round int) (interface{}, error) {
	switch t {
	case TypeNumber:
		return parseFloat64(*(*string)(unsafe.Pointer(&b)), round)
	case TypeBool:
		return strconv.ParseBool(*(*string)(unsafe.Pointer(&b)))
	case TypeString:
		return *(*string)(unsafe.Pointer(&b)), nil
	default:
		return nil, fmt.Errorf("unknow type %s", t)
	}
}

// ParseSize parses a human readble byte string data size notation like `10.5M`
// to the given unit b/kb/mb/gb/tb/pb/eb/kib/mib/gib/tib/pib/eib
func ParseSize(b []byte, unit Unit) (float64, error) {
	// Prepare string and match against the given string

	b = bytes.TrimSpace(b)
	b = bytes.ToLower(b)

	match := rexParseSize.FindSubmatch(b)
	if match == nil {
		return 0, fmt.Errorf("rexon: could not match string for parsing: %s", string(b))
	}

	// Parse the number part to float
	value, err := parseFloat64(*(*string)(unsafe.Pointer(&match[1])), -1)
	if err != nil {
		return 0, fmt.Errorf("rexon: %s", err)
	}
	valueUnit := Unit(value)

	// Convert from the original unit to bytes
	switch string(match[2]) {
	case "", "b", "byte":
	// nothing to do
	case "k", "kb", "kilo", "kilobyte", "kilobytes":
		valueUnit = valueUnit * KBytes
	case "m", "mb", "mega", "megabyte", "megabytes":
		valueUnit = valueUnit * MBytes
	case "g", "gb", "giga", "gigabyte", "gigabytes":
		valueUnit = valueUnit * GBytes
	case "t", "tb", "tera", "terabyte", "terabytes":
		valueUnit = valueUnit * TBytes
	case "p", "pb", "peta", "petabyte", "petabytes":
		valueUnit = valueUnit * PBytes
	case "e", "eb", "exa", "exabyte", "exabytes":
		valueUnit = valueUnit * EBytes
	case "ki", "kib", "kibibyte", "kibibytes":
		valueUnit = valueUnit * KiBytes
	case "mi", "mib", "mebibyte", "mebibytes":
		valueUnit = valueUnit * MiBytes
	case "gi", "gib", "gibibyte", "gibibytes":
		valueUnit = valueUnit * GiBytes
	case "ti", "tib", "tebibyte", "tebibytes":
		valueUnit = valueUnit * TiBytes
	case "pi", "pib", "pebibyte", "pebibytes":
		valueUnit = valueUnit * PiBytes
	case "ei", "eib", "exbibyte", "exbibytes":
		valueUnit = valueUnit * EiBytes
	default:
		return 0, fmt.Errorf("rexon: cannot parse from unit: %s", match[2])
	}
	// Return the parsed value in the specified unit
	return float64(valueUnit / unit), nil
}

// ParseSizeString is like ParseSize but accepts a string argument
func ParseSizeString(s string, unit Unit) (float64, error) {
	shdr := *(*reflect.StringHeader)(unsafe.Pointer(&s))
	bhdr := reflect.SliceHeader{Data: shdr.Data, Len: shdr.Len, Cap: shdr.Len}
	bv := *(*[]byte)(unsafe.Pointer(&bhdr))
	return ParseSize(bv, unit)
}

// parseFloat64 parses a string into a float64 rounding it to the round precision
func parseFloat64(s string, r int) (float64, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}

	if r < 0 {
		return f, nil
	}
	return Round(f, r), nil
}

// Round a float to the specified precision
func Round(f float64, round int) float64 {
	shift := math.Pow(10, float64(round))
	return math.Floor((f*shift)+.5) / shift
}
