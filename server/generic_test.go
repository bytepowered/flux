package server

import (
	"github.com/bytepowered/flux"
	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCopierMap(t *testing.T) {
	var src flux.ServiceArgumentSpec
	(&src).SetExtends("abc", "ABC")
	var dst flux.ServiceArgumentSpec
	(&dst).SetExtends("123", "123")
	(&dst).SetExtends("abc", "123")

	// 合并，保留
	_ = copier.Copy(&dst, &src)
	tester := assert.New(t)
	vOfKey123, ok := (&dst).GetExtends("123")
	tester.True(ok, "should be ok: dist argument, origin values must exists")
	tester.Equal("123", vOfKey123)
	vOfKabc, _ := (&dst).GetExtends("abc")
	tester.Equal("123", vOfKabc)

	// 修改
	(&dst).SetExtends("abc", "456")
	iv1, ok1 := (&src).GetExtends("abc")
	tester.True(ok1, "should be ok: src argument, origin values must not modified")
	tester.Equal("ABC", iv1)
}
