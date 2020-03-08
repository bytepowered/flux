package pkg

import (
	"math"
	"testing"
)

var (
	toIntValues = []struct {
		input    interface{}
		expected int
	}{
		{int(2020), 2020},
		{int8(20), 20},
		{int16(2020), 2020},
		{int32(2020), 2020},
		{int64(2020), 2020},
		{uint(2020), 2020},
		{uint8(20), 20},
		{uint16(2020), 2020},
		{uint32(2020), 2020},
		{uint64(2020), 2020},
		{float32(2020.123), 2020},
		{float64(2020.456), 2020},
		{"2020", 2020},
		{"2020.123", 2020},
	}

	toFloatValues = []struct {
		input    interface{}
		expected float32
	}{
		{int(2020), 2020},
		{int8(20), 20},
		{int16(2020), 2020},
		{int32(2020), 2020},
		{int64(2020), 2020},
		{uint(2020), 2020},
		{uint8(20), 20},
		{uint16(2020), 2020},
		{uint32(2020), 2020},
		{uint64(2020), 2020},
		{float32(2020.123), 2020.123},
		{float64(2020.456), 2020.456},
		{"2020", 2020},
		{"2020.123", 2020.123},
	}
)

func TestSupportedValueToInt(t *testing.T) {
	for _, ca := range toIntValues {
		if v, e := ToInt(ca.input); nil != e {
			t.Errorf("Error, input= %+v, output=%d, error: %s", ca.input, v, e)
		} else if v != ca.expected {
			t.Errorf("NotMatch, input=%+v, output=%d, expected=%d", ca.input, v, ca.expected)
		}
	}
}

func TestNotSupportedValueToInt(t *testing.T) {
	v, e := ToInt(struct {
	}{})
	if nil == e {
		t.Errorf("ToInt should return error, was, value: %d", v)
	} else {
		t.Logf("ToInt should return error: %s", e)
	}
}

func BenchmarkToInt(b *testing.B) {
	// goos: darwin
	// goarch: amd64
	// BenchmarkToInt-4   	68758084	        15.3 ns/op
	for i := 0; i < b.N; i++ {
		_, _ = ToInt64(toIntValues[i%len(toIntValues)].input)
	}
}

func TestSupportedValueToFloat(t *testing.T) {
	for _, ca := range toFloatValues {
		if v, e := ToFloat32(ca.input); nil != e {
			t.Errorf("Error, input= %+v, output=%f, error: %s", ca.input, v, e)
		} else if math.Dim(float64(v), float64(ca.expected)) > 0.001 {
			t.Errorf("NotMatch, input=%+v, output=%f, expected=%f", ca.input, v, ca.expected)
		}
	}
}

func TestNotSupportedValueToFloat(t *testing.T) {
	v, e := ToFloat32(struct {
	}{})
	if nil == e {
		t.Errorf("ToFloat should return error, was, value: %f", v)
	} else {
		t.Logf("ToFloat should return error: %s", e)
	}
}

func BenchmarkToFloat(b *testing.B) {
	// goos: darwin
	// goarch: amd64
	// BenchmarkToFloat-4   	60925686	        19.1 ns/op
	for i := 0; i < b.N; i++ {
		_, _ = ToFloat64(toFloatValues[i%len(toIntValues)].input)
	}
}
