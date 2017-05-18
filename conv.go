package rexon

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	// TypeInt identifies in RexLine/Set.FieldTypes that the given field should be parsed to int
	TypeInt ValueType = iota
	// TypeFloat identifies in RexLine/Set.FieldTypes that the given field should be parsed to float
	TypeFloat ValueType = iota
	// TypeBool identifies in RexLine/Set.FieldTypes that the given field should be parsed to bool
	TypeBool ValueType = iota
	// TypeString identifies in RexLine/Set.FieldTypes that the given field should be parsed to string
	TypeString ValueType = iota
	// Bytes bytes
	Bytes = 1
	// KBytes KiloBytes
	KBytes = Bytes * 1000
	// MBytes MegaBytes
	MBytes = KBytes * 1000
	// GBytes GigaBytes
	GBytes = MBytes * 1000
	// TBytes TeraBytes
	TBytes = GBytes * 1000
	// PBytes PetaBytes
	PBytes = TBytes * 1000
	// EBytes ExaBytes
	EBytes = PBytes * 1000
	// KiBytes KiloBytes
	KiBytes = Bytes * 1024
	// MiBytes MegaBytes
	MiBytes = KiBytes * 1024
	// GiBytes GigaBytes
	GiBytes = MiBytes * 1024
	// TiBytes TeraBytes
	TiBytes = GBytes * 1024
	// PiBytes PetaBytes
	PiBytes = TBytes * 1024
	// EiBytes ExaBytes
	EiBytes = PBytes * 1024
)

// ParseBool parses a string to a bool
func ParseBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}

// ParseInt64 parses a string into a int64
func ParseInt64(s string) (int64, error) {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}

// ParseFloat64 parses a string into a float64 rounding it to the round precision
func ParseFloat64(s string, round int) (float64, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return Round(f, round), nil
}

// Round a float to the specified precision
func Round(f float64, round int) float64 {
	shift := math.Pow(10, float64(round))
	return math.Floor((f*shift)+.5) / shift
}

// func round(v float64, decimals int) float64 {
//      var pow float64 = 1
//      for i:=0; i<decimals; i++ {
//          pow *= 10
//      }
//      return float64(int((v * pow) + 0.5)) / pow
// }

// ParseValue the given []byte field to the specified ValueType.
// The round argument is only used when the ValueType is TypeFloat.
func ParseValue(b []byte, t ValueType, round int) (interface{}, error) {
	switch t {
	case TypeInt:
		return ParseInt64(string(b))
	case TypeFloat:
		return ParseFloat64(string(b), round)
	case TypeBool:
		return ParseBool(string(b))
	case TypeString:
		return b, nil
	default:
		return nil, fmt.Errorf("rexon: could not match type: %s for: %s", t, b)
	}
}

// ParseJsonValues parses values in a []byte encoded json document to the type specified type
// for each key in the fieldTypes argument. Ommited fields will resort to the KeyTypeAll type
// if specified, else the current value and type will remain unaltered.
func ParseJsonValues(b []byte, fieldTypes map[string]ValueType, round int) ([]byte, error) {
	var err error
	var parsedValue interface{}
	result := gjson.ParseBytes(b)

	// Iterate over document key and values for
	result.ForEach(func(key, value gjson.Result) bool {
		// Get the type for field
		valueType, ok := fieldTypes[key.String()]
		if !ok {
			// Try to use the catch all type
			if all, ok := fieldTypes[KeyTypeAll]; ok {
				valueType = all
			} else {
				return true
			}
		}

		if parsedValue, err = ParseValue([]byte(value.String()), valueType, round); err != nil {
			return false
		}

		if b, err = sjson.SetBytes(b, key.String(), parsedValue); err != nil {
			return false
		}
		return true
	})
	return b, err
}

// ParseSize parses a human readble string data size notation like `10.5M`
// to the given unit b/kb/mb/gb/tb/pb/eb/kib/mib/gib/tib/pib/eib
func ParseSize(str string, unit float64) (float64, error) {
	// Prepare string and match against the given string
	str = strings.TrimSpace(str)
	str = strings.ToLower(str)

	match := rexParseSize.FindStringSubmatch(str)
	if match == nil {
		return 0, fmt.Errorf("rexon: could not match string for parsing: %s", str)
	}

	// Parse the number part to float
	value, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, err
	}

	// Convert from the original unit to bytes
	switch match[2] {
	case "", "b", "byte":
	// nothing to do
	case "k", "kb", "kilo", "kilobyte", "kilobytes":
		value = value * KBytes
	case "m", "mb", "mega", "megabyte", "megabytes":
		value = value * MBytes
	case "g", "gb", "giga", "gigabyte", "gigabytes":
		value = value * GBytes
	case "t", "tb", "tera", "terabyte", "terabytes":
		value = value * TBytes
	case "p", "pb", "peta", "petabyte", "petabytes":
		value = value * PBytes
	case "e", "eb", "exa", "exabyte", "exabytes":
		value = value * EBytes
	case "ki", "kib", "kibibyte", "kibibytes":
		value = value * KiBytes
	case "mi", "mib", "mebibyte", "mebibytes":
		value = value * MiBytes
	case "gi", "gib", "gibibyte", "gibibytes":
		value = value * GiBytes
	case "ti", "tib", "tebibyte", "tebibytes":
		value = value * TiBytes
	case "pi", "pib", "pebibyte", "pebibytes":
		value = value * PiBytes
	case "ei", "eib", "exbibyte", "exbibytes":
		value = value * EiBytes
	default:
		return 0, fmt.Errorf("rexon: could not parse from unit: %s", match[2])
	}
	// Return the parsed value in the specified unit
	return value / unit, nil
}
