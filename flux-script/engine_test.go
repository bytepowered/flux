package fluxscript

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"sync"
	"testing"
)

func TestEvalScript(t *testing.T) {
	se := NewEngine()
	asserter := assert.New(t)
	// Eval
	for i := 0; i < 100; i++ {
		v, err := se.EvalEntry(`
function entry(ctx) {
	return "HelloWorld";
}
`, ScriptContext{})
		asserter.Nil(err, "eval: error must nil")
		asserter.Equal("HelloWorld", v, "eval: return value must match")
	}
}

// goos: darwin
// goarch: amd64
// MacBook Pro (13-inch, 2017) BenchmarkEvalScript-4   	   46208	     32371 ns/op
func BenchmarkEvalScript(b *testing.B) {
	se := NewEngine()
	asserter := assert.New(b)
	ctx := ScriptContext{}
	for i := 0; i < b.N; i++ {
		v, err := se.EvalEntry(`
function entry(ctx) {
	return "HelloWorld";
}
`, ctx)
		asserter.Nil(err, "eval: error must nil")
		asserter.Equal("HelloWorld", v, "eval: return value must match")
	}
}

func TestLoadScriptId(t *testing.T) {
	se := NewEngine()
	id, err := se.Load(`
function entry(ctx) {
	return ctx.getFormVar("key");
}
`)
	asserter := assert.New(t)
	asserter.Nil(err, "load: error must nil")
	asserter.NotEmpty(id, "load: id must not empty")
	println("script.id=" + id)
	// Eval
	wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		go func(idx int) {
			wg.Add(1)
			defer wg.Done()
			ctx := ScriptContext{
				GetFormVarFunc: func(key string) string {
					return strconv.Itoa(idx)
				},
			}
			v, err := se.EvalEntryScriptId(id, ctx)
			asserter.Nil(err, "eval: error must nil")
			asserter.Equal(strconv.Itoa(idx), v, "eval: return value must match")
		}(i)
	}
	wg.Wait()
}

// goos: darwin
// goarch: amd64
// MacBook Pro (13-inch, 2017) BenchmarkLoadScriptId-4   	   29760	     41278 ns/op
func BenchmarkLoadScriptId(b *testing.B) {
	se := NewEngine()
	id, err := se.Load(`
function entry(ctx) {
	var rnd = ctx.random(100);
	for (var i = 0; i < 100; i++) {
		rnd += i;
	}
	var key = rnd + "";
	return ctx.getFormVar(key);
}
`)
	asserter := assert.New(b)
	asserter.Nil(err, "load: error must nil")
	asserter.NotEmpty(id, "load: id must not empty")
	ctx := ScriptContext{
		RandomInt63Func: func(max int64) int64 {
			return rand.Int63n(max)
		},
		GetFormVarFunc: func(key string) string {
			return "HelloWorld"
		},
	}
	for i := 0; i < b.N; i++ {
		v, err := se.EvalEntryScriptId(id, ctx)
		asserter.Nil(err, "eval: error must nil")
		asserter.Equal("HelloWorld", v, "eval: return value must match")
	}
}
