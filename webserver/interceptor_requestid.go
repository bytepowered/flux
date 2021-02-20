package webserver

import (
	"github.com/bwmarrin/snowflake"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
)

// RequestIdLookupFunc 查找或者生成RequestId的函数
type RequestIdLookupFunc func(ctx flux.WebContext) string

var (
	defaultLookupHeaders = map[string]struct{}{
		flux.HeaderXRequestId: {},
		flux.HeaderXRequestID: {},
		"requestId":           {},
		"request-id":          {},
	}
)

// NewRequestIdInterceptorHeaders 生成从Header中查找的RequestId中间件的函数
func NewRequestIdInterceptorHeaders(headers ...string) flux.WebInterceptor {
	id, err := snowflake.NewNode(1)
	if nil != err {
		logger.Panicw("request-id-middleware: new snowflake node", "error", err)
		return nil
	}
	for _, name := range headers {
		defaultLookupHeaders[name] = struct{}{}
	}
	names := make([]string, 0)
	for name := range defaultLookupHeaders {
		names = append(names, name)
	}
	return NewRequestIdInterceptor(NewRequestIdLookupFunc(names, id))
}

// NewRequestIdInterceptor 生成RequestId中间件的函数
func NewRequestIdInterceptor(lookup RequestIdLookupFunc) flux.WebInterceptor {
	return func(next flux.WebHandler) flux.WebHandler {
		return func(webc flux.WebContext) error {
			id := lookup(webc)
			webc.SetVariable(flux.HeaderXRequestId, id)
			webc.SetResponseHeader(flux.HeaderXRequestId, id)
			return next(webc)
		}
	}
}

func NewRequestIdLookupFunc(names []string, generator *snowflake.Node) RequestIdLookupFunc {
	return func(webc flux.WebContext) string {
		// 查Header
		for _, name := range names {
			id := webc.HeaderVar(name)
			if "" != id {
				return id
			}
		}
		// 生成随机Id
		return generator.Generate().Base64()
	}
}
