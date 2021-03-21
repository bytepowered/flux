package fluxscript

import (
	"github.com/stretchr/testify/assert"
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
// BenchmarkEvalScript
// BenchmarkEvalScript-12    	    8240	    145634 ns/op
func BenchmarkEvalScript(b *testing.B) {
	se := NewEngine()
	asserter := assert.New(b)
	for i := 0; i < b.N; i++ {
		v, err := se.EvalEntry(`
function entry(ctx) {
	return "HelloWorld";
}
`, ScriptContext{})
		asserter.Nil(err, "eval: error must nil")
		asserter.Equal("HelloWorld", v, "eval: return value must match")
	}
}

func TestLoadScriptId(t *testing.T) {
	se := NewEngine()
	id, err := se.Load(`
function entry(ctx) {
	return "HelloWorld";
}
`)
	asserter := assert.New(t)
	asserter.Nil(err, "load: error must nil")
	asserter.NotEmpty(id, "load: id must not empty")
	println("script.id=" + id)
	// Eval
	for i := 0; i < 100; i++ {
		v, err := se.EvalEntryScriptId(id, ScriptContext{})
		asserter.Nil(err, "eval: error must nil")
		asserter.Equal("HelloWorld", v, "eval: return value must match")
	}
}

// goos: darwin
// goarch: amd64
// BenchmarkLoadScriptId
// BenchmarkLoadScriptId-12    	    8259	    139677 ns/op
func BenchmarkLoadScriptId(b *testing.B) {
	se := NewEngine()
	id, err := se.Load(`
function entry(ctx) {
	return "HelloWorld";
}
`)
	asserter := assert.New(b)
	asserter.Nil(err, "load: error must nil")
	asserter.NotEmpty(id, "load: id must not empty")
	for i := 0; i < b.N; i++ {
		v, err := se.EvalEntryScriptId(id, ScriptContext{})
		asserter.Nil(err, "eval: error must nil")
		asserter.Equal("HelloWorld", v, "eval: return value must match")
	}
}
