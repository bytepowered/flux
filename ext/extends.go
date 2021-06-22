package ext

import "github.com/bytepowered/flux"

type extkey struct {
	id string
}

// ArgumentValueLoaderFunc 参数值直接加载函数
type (
	ArgumentValueLoaderFunc func() flux.MTValue
)

var (
	extkeyValueLoader = extkey{id: "argument.value.loader.func"}
)

func SetArgumentValueLoader(arg *flux.ServiceArgumentSpec, f ArgumentValueLoaderFunc) {
	arg.SetExtends(extkeyValueLoader, f)
}

func GetArgumentValueLoader(arg *flux.ServiceArgumentSpec) (ArgumentValueLoaderFunc, bool) {
	v, ok := arg.GetExtends(extkeyValueLoader)
	if ok {
		f, is := v.(ArgumentValueLoaderFunc)
		return f, is
	}
	return nil, false
}
