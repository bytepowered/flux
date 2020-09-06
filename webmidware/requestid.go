package webmidware

import (
	"github.com/bwmarrin/snowflake"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
)

// LookupRequestIdFunc 查找或者生成RequestId的函数
type LookupRequestIdFunc func(ctx flux.WebContext) string

var (
	_defaultLookupHeaders = map[string]struct{}{
		flux.HeaderXRequestId: {},
		flux.HeaderXRequestID: {},
		"requestId":           {},
		"request-id":          {},
	}
	_lookupRequestIdFunc LookupRequestIdFunc
)

// AddRequestIdLookupHeader 添加默认查找RequestId的Header名称。
// 注意：在注册RequestIdMiddleware前添加生效
func AddRequestIdLookupHeader(header string) {
	_defaultLookupHeaders[header] = struct{}{}
}

// SetRequestIdLookupFunc 设置查找RequestId的函数
// 注意：在注册RequestIdMiddleware前添加生效
func SetRequestIdLookupFunc(f LookupRequestIdFunc) {
	_lookupRequestIdFunc = f
}

// NewRequestIdMiddlewareWithinHeader 生成从Header中查找的RequestId中间件的函数
func NewRequestIdMiddlewareWithinHeader(headers ...string) flux.WebMiddleware {
	id, err := snowflake.NewNode(1)
	if nil != err {
		logger.Panicw("request-id-middleware: new snowflake node", "error", err)
		return nil
	}
	for _, name := range headers {
		AddRequestIdLookupHeader(name)
	}
	names := make([]string, 0)
	for name := range _defaultLookupHeaders {
		names = append(names, name)
	}
	return NewRequestIdMiddleware(DefaultLookupRequestIdFactory(names, id))
}

// NewRequestIdMiddleware 生成RequestId中间件的函数
func NewRequestIdMiddleware(lookupFunc LookupRequestIdFunc) flux.WebMiddleware {
	return func(next flux.WebRouteHandler) flux.WebRouteHandler {
		return func(webc flux.WebContext) error {
			requestId := lookupFunc(webc)
			webc.SetValue(flux.HeaderXRequestId, requestId)
			webc.SetRequestHeader(flux.HeaderXRequestId, requestId)
			webc.SetResponseHeader(flux.HeaderXRequestId, requestId)
			return next(webc)
		}
	}
}

func DefaultLookupRequestIdFactory(names []string, generator *snowflake.Node) LookupRequestIdFunc {
	return func(webc flux.WebContext) string {
		// 查指定查找函数
		if nil != _lookupRequestIdFunc {
			id := _lookupRequestIdFunc(webc)
			if id != "" {
				return id
			}
		}
		// 查Header
		for _, name := range names {
			id := webc.GetRequestHeader(name)
			if "" != id {
				return id
			}
		}
		// 生成随机Id
		return generator.Generate().Base64()
	}
}
