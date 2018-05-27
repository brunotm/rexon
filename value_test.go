package rexon

import (
	"reflect"
	"testing"
)

var (
	dataString   = []byte(`string: 45j45h45kh5hbbbb`)
	dataNumber   = []byte(`number: 1.6445`)
	dataDigital  = []byte(`digital: 1.6445666GB`)
	dataTime     = []byte(`time: 2018-12-25 15:04:05`)
	dataDuration = []byte(`duration: 5m`)
)

func TestValueParseString(t *testing.T) {
	v, err := NewValue(
		"string",
		String,
		ValueRegex(`string:\s+(\w+)`))

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
	v, err := NewValue(
		"number",
		Number,
		Round(2),
		ValueRegex(`number:\s+([-+]?\d*\.?\d+)`))

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
	v, err := NewValue(
		"digital",
		Number,
		FromFormat("digital_unit"),
		ToFormat("mb"),
		Round(3),
		ValueRegex(`digital:\s+([-+]?\d*\.?\d+\w*)`))

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
	v, err := NewValue(
		"time",
		Time,
		ToFormat("rfc3339"),
		FromFormat("2006-01-02 15:04:05"),
		ValueRegex(`time:\s+(.*)`))

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
	v, err := NewValue(
		"duration",
		Duration,
		ToFormat("sec"),
		ValueRegex(`duration:\s+(.*)`))

	if err != nil {
		t.Fatal(err)
	}
	value, ok, err := v.Parse(dataDuration)
	if !ok || err != nil {
		t.Fatal(ok, err)
	}
	t.Log("value: ", reflect.TypeOf(value), value)
}
