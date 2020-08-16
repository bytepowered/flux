package middleware

import (
	"github.com/bwmarrin/snowflake"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/webx"
)

var (
	_defaultLookupHeaders = map[string]struct{}{
		webx.HeaderXRequestId: {},
		webx.HeaderXRequestID: {},
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
type LookupRequestIdFunc func(ctx webx.WebContext) string

// NewRequestIdMiddleware 生成RequestId中间件的函数
func NewRequestIdMiddleware(headers ...string) webx.WebMiddleware {
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
	return NewLookupRequestIdMiddleware(AutoGenerateRequestIdFactory(names, id))
}

// NewRequestIdMiddleware 生成RequestId中间件的函数
func NewLookupRequestIdMiddleware(lookupFunc LookupRequestIdFunc) webx.WebMiddleware {
	return func(next webx.WebRouteHandler) webx.WebRouteHandler {
		return func(webc webx.WebContext) error {
			requestId := lookupFunc(webc)
			webc.SetValue(webx.HeaderXRequestId, requestId)
			webc.RequestHeader().Set(webx.HeaderXRequestId, requestId)
			webc.ResponseHeader().Set(webx.HeaderXRequestId, requestId)
			return next(webc)
		}
	}
}

func AutoGenerateRequestIdFactory(names []string, generator *snowflake.Node) LookupRequestIdFunc {
	return func(webc webx.WebContext) string {
		header := webc.RequestHeader()
		for _, name := range names {
			id := header.Get(name)
			if "" != id {
				return id
			}
		}
		return generator.Generate().Base64()
	}
}
