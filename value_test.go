package rexon

import (
	"reflect"
	"testing"
)

var (
	dataString      = []byte(`string: 45j45h45kh5hbbbb`)
	dataValueString = []byte(`{
		"type": "string",
		"regexp": "string:\\s+(\\w+)"
	}`)

	dataNumber      = []byte(`number: 1.6445`)
	dataValueNumber = []byte(`{
		"type": "number",
		"round": 2,
		"regexp": "number:\\s+([-+]?\\d*\\.?\\d+)"
	}`)

	dataDigital      = []byte(`digital: 1.6445666GB`)
	dataValueDigital = []byte(`{
		"type":  "number",
		"from_format": "digital_unit",
		"to_format": "mb",
		"round": 3,
		"regexp": "digital:\\s+([-+]?\\d*\\.?\\d+\\w*)"
	}`)

	dataTime      = []byte(`time: 2018-12-25 15:04:05`)
	dataValueTime = []byte(`{
		"type": "time",
		"to_format": "rfc3339",
		"from_format": "2006-01-02 15:04:05",
		"regexp": "time:\\s+(.*)"
	}`)

	dataDuration      = []byte(`duration: 5m`)
	dataValueDuration = []byte(`{
		"type": "duration",
		"to_format": "sec",
		"regexp": "duration:\\s+(.*)"
	}`)
)

func valueUnmarshal(data []byte) (v *ValueParser, err error) {
	v = &ValueParser{}
	err = v.UnmarshalJSON(data)
	return v, err
}

func TestValueUnmarshal(t *testing.T) {
	_, err := valueUnmarshal(dataValueNumber)
	if err != nil {
		t.Fatal(err)
	}
}

func TestValueParseString(t *testing.T) {
	v, err := valueUnmarshal(dataValueString)
	if err != nil {
		t.Fatal(err)
	}
	value, ok, err := v.Parse(dataString)
	if !ok || err != nil {
		t.Fatal(ok, err)
	}
	t.Log("value: ", reflect.TypeOf(value), value)
}

func TestValueParseNumber(t *testing.T) {
	v, err := valueUnmarshal(dataValueNumber)
	if err != nil {
		t.Fatal(err)
	}
	value, ok, err := v.Parse(dataNumber)
	if !ok || err != nil {
		t.Fatal(ok, err)
	}
	t.Log("value: ", reflect.TypeOf(value), value)
}

func TestValueParseDigital(t *testing.T) {
	v, err := valueUnmarshal(dataValueDigital)
	if err != nil {
		t.Fatal(err)
	}
	value, ok, err := v.Parse(dataDigital)
	if !ok || err != nil {
		t.Fatal(ok, err)
	}
	t.Log("value: ", reflect.TypeOf(value), value)
}

func TestValueParseTime(t *testing.T) {
	v, err := valueUnmarshal(dataValueTime)
	if err != nil {
		t.Fatal(err)
	}
	value, ok, err := v.Parse(dataTime)
	if !ok || err != nil {
		t.Fatal(ok, err)
	}
	t.Log("value: ", reflect.TypeOf(value), value)
}

func TestValueParseDuration(t *testing.T) {
	v, err := valueUnmarshal(dataValueDuration)
	if err != nil {
		t.Fatal(err)
	}
	value, ok, err := v.Parse(dataDuration)
	if !ok || err != nil {
		t.Fatal(ok, err)
	}
	t.Log("value: ", reflect.TypeOf(value), value)
}
