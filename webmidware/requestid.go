package webmidware

import (
	"github.com/bwmarrin/snowflake"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
)

var (
	_defaultLookupHeaders = map[string]struct{}{
		flux.HeaderXRequestId: {},
		flux.HeaderXRequestID: {},
		"requestId":           {},
		"request-id":          {},
	}
)

// AddRequestIdLookupHeader 添加查找RequestId的Header名称。
// 注意：在注册RequestIdMiddleware前添加生效
func AddRequestIdLookupHeader(header string) {
	_defaultLookupHeaders[header] = struct{}{}
}

// LookupRequestIdFunc 查找或者生成RequestId的函数
type LookupRequestIdFunc func(ctx flux.WebContext) string

// NewRequestIdMiddleware 生成RequestId中间件的函数
func NewRequestIdMiddleware(headers ...string) flux.WebMiddleware {
	id, err := snowflake.NewNode(1)
	if nil != err {
		logger.Panicw("request-id-webmidware: new snowflake node", "error", err)
		return nil
	}
	for _, name := range headers {
		AddRequestIdLookupHeader(name)
	}
	names := make([]string, 0)
	for name := range _defaultLookupHeaders {
		names = append(names, name)
	}
	return NewLookupRequestIdMiddleware(AutoGenerateRequestIdFactory(names, id))
}

// NewRequestIdMiddleware 生成RequestId中间件的函数
func NewLookupRequestIdMiddleware(lookupFunc LookupRequestIdFunc) flux.WebMiddleware {
	return func(next flux.WebRouteHandler) flux.WebRouteHandler {
		return func(webc flux.WebContext) error {
			requestId := lookupFunc(webc)
			webc.SetValue(flux.HeaderXRequestId, requestId)
			webc.SetRequestHeader(flux.HeaderXRequestId, requestId)
			webc.SetRequestHeader(flux.HeaderXRequestId, requestId)
			return next(webc)
		}
	}
}

func AutoGenerateRequestIdFactory(names []string, generator *snowflake.Node) LookupRequestIdFunc {
	return func(webc flux.WebContext) string {
		for _, name := range names {
			id := webc.GetRequestHeader(name)
			if "" != id {
				return id
			}
		}
		return generator.Generate().Base64()
	}
}
