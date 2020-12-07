package support

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	assert2 "github.com/stretchr/testify/assert"
)

func Test_QueryToJsonBytes(t *testing.T) {
	query := `foo=bar&abc=001&foo=value&data="abc"`
	jb, err := JSONBytesFromQueryString([]byte(query))
	assert := assert2.New(t)
	assert.NoError(err, "parse, must not error")
	jbs := string(jb)
	fmt.Println(jbs)
	hmap := make(map[string]interface{})
	err = json.Unmarshal(jb, &hmap)
	assert.NoError(err, "Unmarshal to map, mast not error")
	assert.Equal([]interface{}{"bar", "value"}, hmap["foo"])
	assert.Equal("001", hmap["abc"])
	assert.Equal(`"abc"`, hmap["data"])
	fmt.Println(hmap)
}

func Benchmark_QueryToJsonBytes(b *testing.B) {
	query := `foo=bar&abc=001&foo=value&data="abc"`
	for i := 0; i < b.N; i++ {
		_, _ = JSONBytesFromQueryString([]byte(query))
	}
}

//// ArrayList

func TestValueToArrayList_Int(t *testing.T) {
	a1, err := CastDecodeMIMEToSliceList([]string{"int"}, flux.MIMEValue{Value: "123", MIMEType: "text"})
	assert := assert2.New(t)
	assert.NoError(err)
	fmt.Println(a1)
	assert.Equal([]interface{}{123}, a1)
}

func TestValueToArrayList_String(t *testing.T) {
	a1, err := CastDecodeMIMEToSliceList([]string{"string"}, flux.MIMEValue{Value: "123", MIMEType: "text"})
	assert := assert2.New(t)
	assert.NoError(err)
	fmt.Println(a1)
	assert.Equal([]interface{}{"123"}, a1)
}

//// StringMap

func TestCastToStringMapUnsupportedError(t *testing.T) {
	assert := assert2.New(t)
	_, err1 := CastDecodeMIMEToStringMap(flux.MIMEValue{Value: "123", MIMEType: "unknown"})
	assert.Error(err1)
}

func TestCastToStringMap_Text(t *testing.T) {
	ext.StoreSerializer(ext.TypeNameSerializerJson, flux.NewJsonSerializer())
	sm, err := CastDecodeMIMEToStringMap(flux.MIMEValue{Value: `{"k":1,"e":"a"}`, MIMEType: flux.ValueMIMETypeGoText})
	assert := assert2.New(t)
	assert.NoError(err)
	assert.Equal(float64(1), sm["k"])
	assert.Equal("a", sm["e"])
}

func TestCastToStringMap_JSONText(t *testing.T) {
	ext.StoreSerializer(ext.TypeNameSerializerJson, flux.NewJsonSerializer())
	sm, err := CastDecodeMIMEToStringMap(flux.MIMEValue{Value: `{"k":1,"e":"a"}`, MIMEType: "application/json"})
	assert := assert2.New(t)
	assert.NoError(err)
	assert.Equal(float64(1), sm["k"])
	assert.Equal("a", sm["e"])
}

func TestCastToStringMap_JSONBytes(t *testing.T) {
	ext.StoreSerializer(ext.TypeNameSerializerJson, flux.NewJsonSerializer())
	sm, err := CastDecodeMIMEToStringMap(flux.MIMEValue{Value: []byte(`{"k":1,"e":"a"}`), MIMEType: "application/json"})
	assert := assert2.New(t)
	assert.NoError(err)
	assert.Equal(float64(1), sm["k"])
	assert.Equal("a", sm["e"])
}

func TestCastToStringMap_JSONReader(t *testing.T) {
	ext.StoreSerializer(ext.TypeNameSerializerJson, flux.NewJsonSerializer())
	sm, err := CastDecodeMIMEToStringMap(flux.MIMEValue{Value: ioutil.NopCloser(strings.NewReader(`{"k":1,"e":"a"}`)), MIMEType: "application/json"})
	assert := assert2.New(t)
	assert.NoError(err)
	assert.Equal(float64(1), sm["k"])
	assert.Equal("a", sm["e"])
}

func TestCastToStringMap_QueryText(t *testing.T) {
	ext.StoreSerializer(ext.TypeNameSerializerJson, flux.NewJsonSerializer())
	sm, err := CastDecodeMIMEToStringMap(flux.MIMEValue{Value: `k=1&e=a`, MIMEType: "application/x-www-form-urlencoded"})
	assert := assert2.New(t)
	assert.NoError(err)
	assert.Equal("1", sm["k"])
	assert.Equal("a", sm["e"])
}

func TestCastToStringMap_QueryBytes(t *testing.T) {
	ext.StoreSerializer(ext.TypeNameSerializerJson, flux.NewJsonSerializer())
	sm, err := CastDecodeMIMEToStringMap(flux.MIMEValue{Value: []byte(`k=1&e=a`), MIMEType: "application/x-www-form-urlencoded"})
	assert := assert2.New(t)
	assert.NoError(err)
	assert.Equal("1", sm["k"])
	assert.Equal("a", sm["e"])
}

func TestCastToStringMap_QueryReader(t *testing.T) {
	ext.StoreSerializer(ext.TypeNameSerializerJson, flux.NewJsonSerializer())
	sm, err := CastDecodeMIMEToStringMap(flux.MIMEValue{Value: ioutil.NopCloser(strings.NewReader(`k=1&e=a`)), MIMEType: "application/x-www-form-urlencoded"})
	assert := assert2.New(t)
	assert.NoError(err)
	assert.Equal("1", sm["k"])
	assert.Equal("a", sm["e"])
}

func TestCastToStringMap_Object1(t *testing.T) {
	assert := assert2.New(t)
	sm, err := CastDecodeMIMEToStringMap(flux.MIMEValue{Value: map[string]interface{}{"a": 1, "b": "c"}, MIMEType: flux.ValueMIMETypeGoObject})
	assert.NoError(err)
	assert.Equal(1, sm["a"])
	assert.Equal("c", sm["b"])
}

func TestCastToStringMap_Object2(t *testing.T) {
	assert := assert2.New(t)
	sm, err := CastDecodeMIMEToStringMap(flux.MIMEValue{Value: map[interface{}]interface{}{"a": 1, "b": "c"}, MIMEType: flux.ValueMIMETypeGoObject})
	assert.NoError(err)
	assert.Equal(1, sm["a"])
	assert.Equal("c", sm["b"])
}
