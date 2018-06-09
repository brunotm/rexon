package rexon

import (
	"encoding/json"
	"strconv"
	"unsafe"

	"github.com/buger/jsonparser"
)

var (
	nullValue = []byte(`null`)
)

func newJSON() (data []byte) {
	data = make([]byte, 0, 64)
	data = append(data, '{', '}')
	return data
}

func jsonHas(data []byte, path ...string) (exists bool) {
	_, _, _, err := jsonparser.Get(data, path...)
	return err != jsonparser.KeyPathNotFoundError
}

func jsonSet(data []byte, value interface{}, path ...string) (d []byte, err error) {
	buf := make([]byte, 0, 16)

	switch v := value.(type) {
	case nil:
		buf = nullValue
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
	case []byte:
		if buf, err = json.Marshal(*(*string)(unsafe.Pointer(&v))); err != nil {
			return nil, err
		}
	// Also catch strings for escaping
	default:
		if buf, err = json.Marshal(v); err != nil {
			return nil, err
		}
	}
	return jsonparser.Set(data, buf, path...)
}
