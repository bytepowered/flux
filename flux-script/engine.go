package fluxscript

import (
	"fmt"
	"github.com/dop251/goja"
	"reflect"
	"strconv"
	"sync"
)

const (
	ScriptEntryFunName = "entry"
)

var engine = &Engine{
	scripts: sync.Map{},
}

type Engine struct {
	scripts sync.Map
}

func NewEngine() *Engine {
	return engine
}

// Load 将JavaScript脚本编译并缓存；返回执行此脚本的ScriptId；
func (se *Engine) Load(source string) (string, error) {
	id := se.scriptId([]byte(source))
	if _, ok := se.scripts.Load(id); !ok {
		pro, err := goja.Compile(id, source, true)
		if nil != err {
			return "", fmt.Errorf("load to compile script, error: %w", err)
		}
		se.scripts.Store(id, pro)
	}
	return id, nil
}

// Exist 判断ScriptId是否存在。
func (se *Engine) Exist(scriptId string) bool {
	_, ok := se.scripts.Load(scriptId)
	return ok
}

// Remove 删除指定ScriptId的脚本
func (se *Engine) Remove(scriptId string) {
	se.scripts.Delete(scriptId)
}

// EvalScriptId 执行指定ScriptId的脚本，执行指定函数；
func (se *Engine) EvalScriptId(scriptId string, entryFun string, context interface{}) (v interface{}, err error) {
	prop, ok := se.scripts.Load(scriptId)
	if !ok || prop == nil {
		return nil, fmt.Errorf("script not found, script-id: %s", scriptId)
	}
	runtime := goja.New()
	_, rerr := runtime.RunProgram(prop.(*goja.Program))
	if nil != rerr {
		return nil, fmt.Errorf("compile script, error: %w", rerr)
	}
	return se.entry(runtime, entryFun, context)
}

// EvalEntryScriptId 执行指定ScriptId的脚本，执行默认entry函数；
func (se *Engine) EvalEntryScriptId(scriptId string, context interface{}) (v interface{}, err error) {
	return se.EvalScriptId(scriptId, ScriptEntryFunName, context)
}

// EvalEntry 执行JavaScript脚本，执行默认entry函数。脚本被立即执行；
func (se *Engine) EvalEntry(source string, context interface{}) (v interface{}, err error) {
	return se.Eval(source, ScriptEntryFunName, context)
}

// Eval 执行JavaScript脚本，指定执行函数。脚本被立即执行；
func (se *Engine) Eval(src string, entryFun string, context interface{}) (v interface{}, err error) {
	runtime := goja.New()
	_, rerr := runtime.RunScript("dynamic.eval.fun:"+entryFun, src)
	if nil != rerr {
		return nil, fmt.Errorf("compile script, error: %w", rerr)
	}
	return se.entry(runtime, entryFun, context)
}

func (se *Engine) scriptId(src []byte) string {
	return strconv.FormatUint(hash64(src), 16)
}

func (se *Engine) entry(runtime *goja.Runtime, entryFun string, entryContext interface{}) (v interface{}, err error) {
	cv := reflect.ValueOf(entryContext)
	if cv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("ScriptContext MUST be struct, was: %s", cv.Kind().String())
	}
	// entry func
	var entry func(goja.Value) interface{}
	verr := runtime.ExportTo(runtime.Get(entryFun), &entry)
	if verr != nil {
		return nil, fmt.Errorf("bind runtime entry function, error: %w", verr)
	}
	defer func() {
		if r := recover(); nil != r {
			err = fmt.Errorf("executing script, error: %s", r)
		}
	}()
	runtime.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	return entry(runtime.ToValue(entryContext)), nil
}
