package ext

import "github.com/bytepowered/flux"

type extkey struct {
	id string
}

var (
	extkeyValueLoader = extkey{id: "argument.value.loader.func"}
)

func SetArgumentValueLoader(arg *flux.Argument, f flux.MTValueLoaderFunc) {
	arg.SetExtends(extkeyValueLoader, f)
}

func GetArgumentValueLoader(arg *flux.Argument) (flux.MTValueLoaderFunc, bool) {
	v, ok := arg.GetExtends(extkeyValueLoader)
	if ok {
		f, is := v.(flux.MTValueLoaderFunc)
		return f, is
	}
	return nil, false
}
