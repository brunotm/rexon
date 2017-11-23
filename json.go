package rexon

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"unicode/utf8"
	"unsafe"

	"github.com/brunotm/jsonparser"
)

// ValueType for type parsing
type ValueType = jsonparser.ValueType

const (
	TypeNumber   ValueType = jsonparser.Number
	TypeString   ValueType = jsonparser.String
	TypeBool     ValueType = jsonparser.Boolean
	TypeArray    ValueType = jsonparser.Array
	TypeObject   ValueType = jsonparser.Object
	TypeNull     ValueType = jsonparser.Null
	TypeNotExist ValueType = jsonparser.NotExist
	TypeUnknown  ValueType = jsonparser.Unknown
)

var (
	errorTypeNotImplemented = errors.New("rexon: type Not implemented")
)

// JSONGet the []byte for the given path
func JSONGet(data []byte, path ...string) (value []byte, valueType ValueType, err error) {
	value, valueType, _, err = jsonparser.Get(data, path...)
	return value, valueType, err
}

// JSONGetValue returns the parsed value for a given path as an interface
func JSONGetValue(data []byte, path ...string) (value interface{}, valueType ValueType, err error) {
	var buf []byte
	buf, valueType, err = JSONGet(data, path...)
	if err != nil {
		return nil, TypeUnknown, err
	}

	switch valueType {
	case TypeNull:
		value = nil
	case TypeNumber:
		value, err = jsonparser.ParseFloat(buf)
	case TypeString:
		value, err = jsonparser.ParseString(buf)
	case TypeBool:
		value, err = jsonparser.ParseBoolean(buf)
	default:
		value = buf
	}
	return value, valueType, err
}

// JSONExists check if the given path exists
func JSONExists(data []byte, path ...string) (exists bool) {
	_, _, _, err := jsonparser.Get(data, path...)
	return err != jsonparser.KeyPathNotFoundError
}

// JSONGetInt fetch the given path as a int64
func JSONGetInt(data []byte, path ...string) (value int64, err error) {
	return jsonparser.GetInt(data, path...)
}

// JSONGetFloat fetch the given path as a float64
func JSONGetFloat(data []byte, path ...string) (value float64, err error) {
	return jsonparser.GetFloat(data, path...)
}

// JSONGetString fetch the given path as a string
func JSONGetString(data []byte, path ...string) (value string, err error) {
	return jsonparser.GetString(data, path...)
}

// JSONGetUnsafeString fetch the given path as a unsafe string string
func JSONGetUnsafeString(data []byte, path ...string) (value string, err error) {
	return jsonparser.GetUnsafeString(data, path...)
}

// JSONGetBool fetch the given path as a bool
func JSONGetBool(data []byte, path ...string) (value bool, err error) {
	return jsonparser.GetBoolean(data, path...)
}

// JSONDelete deletes the value for the given path
func JSONDelete(data []byte, path ...string) (d []byte) {
	return jsonparser.Delete(data, path...)
}

// JSONForEach applies the given func to every key/value pair in the json data
func JSONForEach(data []byte, cb func(key string, value []byte, tp ValueType) error, keys ...string) (err error) {

	iter := func(key []byte, value []byte, tp ValueType, offset int) error {
		return cb(*(*string)(unsafe.Pointer(&key)), value, tp)
	}
	return jsonparser.ObjectEach(data, iter, keys...)
}

// JSONSetRawBytes set the given []byte value for the provided path
func JSONSetRawBytes(data []byte, value []byte, path ...string) (d []byte, err error) {
	return jsonparser.Set(data, value, path...)
}

// JSONSet parses and set the given value for the provided path.
// byte slices will be set as string, value is already a parsed type use JSONSetRawBytes instead
func JSONSet(data []byte, value interface{}, path ...string) (d []byte, err error) {
	buf := make([]byte, 0, 8)

	// Creates a JSON if we're given a nil or empty []byte
	if len(data) == 0 {
		data = make([]byte, 0, 512)
		data = append(data, '{', '}')
	}

	switch v := value.(type) {
	case []byte:
		buf = marshalByteString(v, false)
	case string:
		shdr := *(*reflect.StringHeader)(unsafe.Pointer(&v))
		bhdr := reflect.SliceHeader{Data: shdr.Data, Len: shdr.Len, Cap: shdr.Len}
		bv := *(*[]byte)(unsafe.Pointer(&bhdr))
		buf = marshalByteString(bv, false)
	case bool:
		buf = strconv.AppendBool(buf, v)
	case int:
		buf = strconv.AppendInt(buf, int64(v), 10)
	case int8:
		buf = strconv.AppendInt(buf, int64(v), 10)
	case int16:
		buf = strconv.AppendInt(buf, int64(v), 10)
	case int32:
		buf = strconv.AppendInt(buf, int64(v), 10)
	case int64:
		buf = strconv.AppendInt(buf, v, 10)
	case uint8:
		buf = strconv.AppendUint(buf, uint64(v), 10)
	case uint16:
		buf = strconv.AppendUint(buf, uint64(v), 10)
	case uint32:
		buf = strconv.AppendUint(buf, uint64(v), 10)
	case uint64:
		buf = strconv.AppendUint(buf, v, 10)
	case float32:
		buf = strconv.AppendFloat(buf, float64(v), 'f', -1, 64)
	case float64:
		buf = strconv.AppendFloat(buf, v, 'f', -1, 64)
	default:
		if buf, err = json.Marshal(value); err != nil {
			return nil, err
		}
	}

	return JSONSetRawBytes(data, buf, path...)
}

// CopyBytes copies a give []byte
func CopyBytes(data []byte) (d []byte) {
	d = make([]byte, len(data))
	copy(d, data)
	return d
}

// String below serialization functionality adapted from encoding/json

const chars = "0123456789abcdef"

func marshalByteString(s []byte, noHTMLEscape bool) []byte {
	buf := make([]byte, 1, len(s)+2)
	buf[0] = '"'

	p := 0 // last non-escape symbol

	for i := 0; i < len(s); {
		c := s[i]

		if isNotEscapedSingleChar(c, !noHTMLEscape) {
			// single-width character, no escaping is required
			i++
			continue
		} else if c < utf8.RuneSelf {
			// single-with character, need to escape
			buf = append(buf, s[p:i]...)
			switch c {
			case '\t':
				buf = append(buf, `\t`...)
			case '\r':
				buf = append(buf, `\r`...)
			case '\n':
				buf = append(buf, `\n`...)
			case '\\':
				buf = append(buf, `\\`...)
			case '"':
				buf = append(buf, `\"`...)
			default:
				buf = append(buf, `\u00`...)
				buf = append(buf, chars[c>>4])
				buf = append(buf, chars[c&0xf])
			}

			i++
			p = i
			continue
		}

		// broken utf
		runeValue, runeWidth := utf8.DecodeRune(s[i:])
		if runeValue == utf8.RuneError && runeWidth == 1 {
			buf = append(buf, s[p:i]...)
			buf = append(buf, `\ufffd`...)
			i++
			p = i
			continue
		}

		// jsonp stuff - tab separator and line separator
		if runeValue == '\u2028' || runeValue == '\u2029' {
			buf = append(buf, s[p:i]...)
			buf = append(buf, `\u202`...)
			buf = append(buf, chars[runeValue&0xf])
			i += runeWidth
			p = i
			continue
		}
		i += runeWidth
	}
	buf = append(buf, s[p:]...)
	buf = append(buf, '"')
	return buf
}

func isNotEscapedSingleChar(c byte, escapeHTML bool) bool {
	// Note: might make sense to use a table if there are more chars to escape. With 4 chars
	// it benchmarks the same.
	if escapeHTML {
		return c != '<' && c != '>' && c != '&' && c != '\\' && c != '"' && c >= 0x20 && c < utf8.RuneSelf
	}
	return c != '\\' && c != '"' && c >= 0x20 && c < utf8.RuneSelf
}
