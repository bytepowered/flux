package common

import (
	"encoding/json"
	"fmt"
	"github.com/bytepowered/flux/flux-node"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/bytepowered/flux/flux-node/ext"
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

var (
	GenericTypeInt    = []string{"int"}
	GenericTypeString = []string{"string"}
)

func TestToGenericList_IntEmpty(t *testing.T) {
	a1, err := ToGenericListE(GenericTypeInt, flux.NewStringMTValue(""))
	assert := assert2.New(t)
	assert.NoError(err)
	fmt.Println(a1)
	assert.Equal([]interface{}{}, a1)
}

func TestToGenericList_IntNil(t *testing.T) {
	a1, err := ToGenericListE(GenericTypeInt, flux.NewObjectMTValue(nil))
	assert := assert2.New(t)
	assert.NoError(err)
	fmt.Println(a1)
	assert.Equal([]interface{}{}, a1)
}

func TestToGenericList_Int(t *testing.T) {
	a1, err := ToGenericListE(GenericTypeInt, flux.NewStringMTValue("123"))
	assert := assert2.New(t)
	assert.NoError(err)
	fmt.Println(a1)
	assert.Equal([]interface{}{123}, a1)
}

func TestToGenericList_IntErr(t *testing.T) {
	_, err := ToGenericListE(GenericTypeInt, flux.NewStringMTValue("abc"))
	assert := assert2.New(t)
	assert.Error(err)
}

func TestToGenericList_String(t *testing.T) {
	a1, err := ToGenericListE(GenericTypeString, flux.NewObjectMTValue(123))
	assert := assert2.New(t)
	assert.NoError(err)
	fmt.Println(a1)
	assert.Equal([]interface{}{"123"}, a1)
}

func TestToGenericList_Nil(t *testing.T) {
	a1, err := ToGenericListE(GenericTypeString, flux.NewObjectMTValue(nil))
	assert := assert2.New(t)
	assert.NoError(err)
	fmt.Println(a1)
	assert.Equal([]interface{}{}, a1)
}

func TestToGenericList_EmptyString(t *testing.T) {
	a1, err := ToGenericListE(GenericTypeString, flux.NewStringMTValue(""))
	assert := assert2.New(t)
	assert.NoError(err)
	fmt.Println(a1)
	assert.Equal([]interface{}{}, a1)
}

func TestToGenericList_ValuesToString(t *testing.T) {
	a1, err := ToGenericListE(GenericTypeString, flux.NewObjectMTValue([]int{123}))
	assert := assert2.New(t)
	assert.NoError(err)
	fmt.Println(a1)
	assert.Equal([]interface{}{"123"}, a1)
}

func TestToGenericList_ValuesToLong(t *testing.T) {
	a1, err := ToGenericListE([]string{"long"}, flux.NewObjectMTValue([]string{"123456"}))
	assert := assert2.New(t)
	assert.NoError(err)
	fmt.Println(a1)
	assert.Equal([]interface{}{int64(123456)}, a1)
}

//// StringMap

func TestToStringMap_Err(t *testing.T) {
	assert := assert2.New(t)
	_, err1 := ToStringMapE(flux.NewStringMTValue("123"))
	assert.Error(err1)
}

func TestToStringMap_Empty(t *testing.T) {
	assert := assert2.New(t)
	sm, err1 := ToStringMapE(flux.NewStringMTValue(""))
	assert.NoError(err1)
	assert.True(0 == len(sm))
}

func TestCastToStringMap_TextEmpty(t *testing.T) {
	ext.RegisterSerializer(ext.TypeNameSerializerJson, flux.NewJsonSerializer())
	sm, err := ToStringMapE(flux.NewStringMTValue(""))
	assert := assert2.New(t)
	assert.NoError(err)
	assert.True(0 == len(sm))
}

func TestCastToStringMap_TextEmptyJSON(t *testing.T) {
	ext.RegisterSerializer(ext.TypeNameSerializerJson, flux.NewJsonSerializer())
	sm, err := ToStringMapE(flux.NewStringMTValue("{}"))
	assert := assert2.New(t)
	assert.NoError(err)
	assert.True(0 == len(sm))
}

func TestCastToStringMap_Text(t *testing.T) {
	ext.RegisterSerializer(ext.TypeNameSerializerJson, flux.NewJsonSerializer())
	sm, err := ToStringMapE(flux.NewStringMTValue(`{"k":1,"e":"a"}`))
	assert := assert2.New(t)
	assert.NoError(err)
	assert.Equal(float64(1), sm["k"])
	assert.Equal("a", sm["e"])
}

func TestCastToStringMap_JSONText(t *testing.T) {
	ext.RegisterSerializer(ext.TypeNameSerializerJson, flux.NewJsonSerializer())
	sm, err := ToStringMapE(flux.MTValue{Value: `{"k":1,"e":"a"}`, MediaType: "application/json"})
	assert := assert2.New(t)
	assert.NoError(err)
	assert.Equal(float64(1), sm["k"])
	assert.Equal("a", sm["e"])
}

func TestCastToStringMap_JSONBytes(t *testing.T) {
	ext.RegisterSerializer(ext.TypeNameSerializerJson, flux.NewJsonSerializer())
	sm, err := ToStringMapE(flux.MTValue{Value: []byte(`{"k":1,"e":"a"}`), MediaType: "application/json"})
	assert := assert2.New(t)
	assert.NoError(err)
	assert.Equal(float64(1), sm["k"])
	assert.Equal("a", sm["e"])
}

func TestCastToStringMap_JSONReader(t *testing.T) {
	ext.RegisterSerializer(ext.TypeNameSerializerJson, flux.NewJsonSerializer())
	sm, err := ToStringMapE(flux.MTValue{Value: ioutil.NopCloser(strings.NewReader(`{"k":1,"e":"a"}`)), MediaType: "application/json"})
	assert := assert2.New(t)
	assert.NoError(err)
	assert.Equal(float64(1), sm["k"])
	assert.Equal("a", sm["e"])
}

func TestCastToStringMap_QueryText(t *testing.T) {
	ext.RegisterSerializer(ext.TypeNameSerializerJson, flux.NewJsonSerializer())
	sm, err := ToStringMapE(flux.MTValue{Value: `k=1&e=a`, MediaType: "application/x-www-form-urlencoded"})
	assert := assert2.New(t)
	assert.NoError(err)
	assert.Equal("1", sm["k"])
	assert.Equal("a", sm["e"])
}

func TestCastToStringMap_QueryBytes(t *testing.T) {
	ext.RegisterSerializer(ext.TypeNameSerializerJson, flux.NewJsonSerializer())
	sm, err := ToStringMapE(flux.MTValue{Value: []byte(`k=1&e=a`), MediaType: "application/x-www-form-urlencoded"})
	assert := assert2.New(t)
	assert.NoError(err)
	assert.Equal("1", sm["k"])
	assert.Equal("a", sm["e"])
}

func TestCastToStringMap_QueryReader(t *testing.T) {
	ext.RegisterSerializer(ext.TypeNameSerializerJson, flux.NewJsonSerializer())
	sm, err := ToStringMapE(flux.MTValue{Value: ioutil.NopCloser(strings.NewReader(`k=1&e=a`)), MediaType: "application/x-www-form-urlencoded"})
	assert := assert2.New(t)
	assert.NoError(err)
	assert.Equal("1", sm["k"])
	assert.Equal("a", sm["e"])
}

func TestCastToStringMap_Object1(t *testing.T) {
	assert := assert2.New(t)
	sm, err := ToStringMapE(flux.MTValue{Value: map[string]interface{}{"a": 1, "b": "c"}, MediaType: flux.MediaTypeGoObject})
	assert.NoError(err)
	assert.Equal(1, sm["a"])
	assert.Equal("c", sm["b"])
}

func TestCastToStringMap_Object2(t *testing.T) {
	assert := assert2.New(t)
	sm, err := ToStringMapE(flux.MTValue{Value: map[interface{}]interface{}{"a": 1, "b": "c"}, MediaType: flux.MediaTypeGoObject})
	assert.NoError(err)
	assert.Equal(1, sm["a"])
	assert.Equal("c", sm["b"])
}
