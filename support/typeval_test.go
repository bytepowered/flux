package support

import (
	"encoding/json"
	"fmt"
	assert2 "github.com/stretchr/testify/assert"
	"testing"
)

func Test_QueryToJsonBytes(t *testing.T) {
	query := `foo=bar&abc=001&foo=value&data="abc"`
	jb, err := _queryToJsonBytes([]byte(query))
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
		_, _ = _queryToJsonBytes([]byte(query))
	}
}
