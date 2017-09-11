package rexon

import (
	"encoding/json"
	"reflect"
	"strconv"
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

// JSONGet the []byte for the given path
func JSONGet(data []byte, path ...string) ([]byte, ValueType, error) {
	dt, tp, _, err := jsonparser.Get(data, path...)
	return dt, tp, err
}

// JSONGetInt fetch the given path as a int64
func JSONGetInt(data []byte, path ...string) (int64, error) {
	return jsonparser.GetInt(data, path...)
}

// JSONGetFloat fetch the given path as a float64
func JSONGetFloat(data []byte, path ...string) (float64, error) {
	return jsonparser.GetFloat(data, path...)
}

// JSONGetString fetch the given path as a string
func JSONGetString(data []byte, path ...string) (string, error) {
	return jsonparser.GetString(data, path...)
}

// JSONGetUnsafeString fetch the given path as a unsafe string string
func JSONGetUnsafeString(data []byte, path ...string) (string, error) {
	return jsonparser.GetUnsafeString(data, path...)
}

// JSONGetBool fetch the given path as a bool
func JSONGetBool(data []byte, path ...string) (bool, error) {
	return jsonparser.GetBoolean(data, path...)
}

// JSONDelete deletes the value for the given path
func JSONDelete(data []byte, path ...string) []byte {
	return jsonparser.Delete(data, path...)
}

// JSONForEach applies the given func to every key/value pair in the json data
func JSONForEach(data []byte, cb func(key string, value []byte, tp ValueType) error, keys ...string) error {

	iter := func(key []byte, value []byte, tp ValueType, offset int) error {
		return cb(*(*string)(unsafe.Pointer(&key)), value, tp)
	}
	return jsonparser.ObjectEach(data, iter, keys...)
}

// JSONSetRawBytes set the given []byte value for the provided path
func JSONSetRawBytes(data []byte, value []byte, path ...string) ([]byte, error) {
	return jsonparser.Set(data, value, path...)
}

// JSONSet parses and set the given value for the provided path
func JSONSet(data []byte, value interface{}, path ...string) ([]byte, error) {
	var res []byte
	var err error
	var str string

	switch v := value.(type) {
	case []byte:
		return JSONSetRawBytes(data, v, path...)
	case string:
		str = strconv.Quote(v)
	case bool:
		if v {
			str = "true"
		} else {
			str = "false"
		}
	case int:
		str = strconv.FormatInt(int64(v), 10)
	case int8:
		str = strconv.FormatInt(int64(v), 10)
	case int16:
		str = strconv.FormatInt(int64(v), 10)
	case int32:
		str = strconv.FormatInt(int64(v), 10)
	case int64:
		str = strconv.FormatInt(v, 10)
	case uint8:
		str = strconv.FormatUint(uint64(v), 10)
	case uint16:
		str = strconv.FormatUint(uint64(v), 10)
	case uint32:
		str = strconv.FormatUint(uint64(v), 10)
	case uint64:
		str = strconv.FormatUint(v, 10)
	case float32:
		str = strconv.FormatFloat(float64(v), 'f', -1, 64)
	case float64:
		str = strconv.FormatFloat(v, 'f', -1, 64)
	default:
		if res, err = json.Marshal(value); err != nil {
			return nil, err
		}
		return JSONSetRawBytes(data, res, path...)
	}

	strHdr := (*reflect.StringHeader)(unsafe.Pointer(&str))
	byteHdr := (*reflect.SliceHeader)(unsafe.Pointer(&res))
	byteHdr.Data = strHdr.Data
	l := len(str)
	byteHdr.Len = l
	byteHdr.Cap = l

	return JSONSetRawBytes(data, res, path...)
}

// CopyBytes copies a give []byte
func CopyBytes(data []byte) []byte {
	b := make([]byte, len(data))
	copy(b, data)
	return b
}