package flux

import (
	"encoding/json"
	"testing"
)

type _serializeKvPair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

var toUnmarshalJsonData = []byte(`[
		{"name": "user.name", "value": "yongjia.chen"},
		{"name": "user.age",    "value": "18"}
	]`)

var toMarshalJsonData = []_serializeKvPair{
	{Name: "user.name", Value: "yongjia.chen"},
	{Name: "user.age", Value: "18"},
}

func TestJsonCase(t *testing.T) {
	m := map[interface{}]interface{}{
		"isDefault":  true,
		"is_success": false,
		"long":       12345678901234,
		"double":     1234567890.123456,
	}
	serializer := NewJsonSerializer()
	b, err := serializer.Marshal(m)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("==> %s", string(b))
}

func TestJsonMapStandard(t *testing.T) {
	serializer := NewJsonSerializer()
	b, err := serializer.Marshal(map[int]bool{1: false})
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("Marshal: %s", string(b))

	var out interface{}
	if err = serializer.Unmarshal([]byte(`{1:false}`), &out); nil != err {
		t.Fatal(err)
	}
	t.Log(out)
}

func Benchmark_SerializeUnmarshal(b *testing.B) {
	serializer := NewJsonSerializer()
	_BenchmarkUnmarshalWith(b, serializer.Unmarshal)
}

func Benchmark_StdLibUnmarshal(b *testing.B) {
	_BenchmarkUnmarshalWith(b, json.Unmarshal)
}

func Benchmark_StdLibMarshal(b *testing.B) {
	_BenchmarkMarshalWith(b, json.Marshal)
}

func Benchmark_SerializeMarshal(b *testing.B) {
	serializer := NewJsonSerializer()
	_BenchmarkMarshalWith(b, serializer.Marshal)
}

func _BenchmarkUnmarshalWith(b *testing.B, unmarshaler func(data []byte, v interface{}) error) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var pairs []_serializeKvPair
		err := unmarshaler(toUnmarshalJsonData, &pairs)
		if err != nil {
			b.Fatal(err)
		}
		if 2 != len(pairs) {
			b.Fatalf("array size not match, was: %d", len(pairs))
		}
		if "18" != pairs[1].Value {
			b.Fatalf("value not match, was: %s", pairs[1].Value)
		}
	}
}

func _BenchmarkMarshalWith(b *testing.B, marshaler func(v interface{}) ([]byte, error)) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bs, err := marshaler(toMarshalJsonData)
		if err != nil {
			b.Fatal(err)
		}
		if string(bs) != `[{"name":"user.name","value":"yongjia.chen"},{"name":"user.age","value":"18"}]` {
			b.Fatalf("data not match, was: %s", string(bs))
		}
	}
}
