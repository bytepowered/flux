package inapp

import (
	"github.com/bytepowered/fluxgo/pkg/flux"
)

var (
	// Note: NonThreadSafe, 不支持动态添加，需要在提供服务前完成注册
	invokers = make(map[string]InvokeFunc, 16)
)

func init() {
	RegisterInvokeFunc("flux.debug.inapp.Test:echo", newDefaultInAppInvokeFunc())
}

func RegisterInvokeFunc(serviceId string, f InvokeFunc) {
	flux.AssertNotEmpty(serviceId, "<service-id> MUST NOT empty")
	flux.AssertNotNil(f, "<invoke-func> MUST NOT nil")
	invokers[serviceId] = f
}

func LoadInvokeFunc(serviceId string) (InvokeFunc, bool) {
	f, ok := invokers[serviceId]
	if ok {
		return f, true
	}
	return nil, false
}
