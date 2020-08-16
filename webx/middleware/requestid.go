package middleware

import (
	"github.com/bwmarrin/snowflake"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/webx"
)

var (
	_lookupHeaderNames = map[string]struct{}{
		webx.HeaderXRequestId: {},
		"requestId":           {},
		"request-id":          {},
	}
)

// AddRequestIdLookupHeader 添加查找RequestId的Header名称。
// 注意：在注册RequestIdMiddleware前添加生效
func AddRequestIdLookupHeader(header string) {
	_lookupHeaderNames[header] = struct{}{}
}

// LookupRequestIdFunc 查找或者生成RequestId的函数
type LookupRequestIdFunc func(ctx webx.WebContext) string

// NewRequestIdMiddleware 生成RequestId中间件的函数
func NewRequestIdMiddleware() webx.WebMiddleware {
	id, err := snowflake.NewNode(1)
	if nil != err {
		logger.Panicw("request-id-middleware: new snowflake node", "error", err)
		return nil
	}
	return NewLookupRequestIdMiddleware(AutoGenerateRequestIdFactory(id))
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

func AutoGenerateRequestIdFactory(generator *snowflake.Node) LookupRequestIdFunc {
	names := make([]string, 0, len(_lookupHeaderNames))
	for name := range _lookupHeaderNames {
		names = append(names, name)
	}
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
