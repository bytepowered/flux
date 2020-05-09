package flux

import (
	"net/http"
)

const (
	KeyConfigRootExchanges = "Exchanges"
)

// Exchange 实现协议的数据通讯
type Exchange interface {
	// Exchange 完成Http与当前协议的数据交互
	Exchange(ctx Context) *InvokeError
	// Invoke 执行指定目标Endpoint的通讯，返回响应结果
	Invoke(target *Endpoint, reqCtx Context) (interface{}, *InvokeError)
}

// ExchangeDecoder 解析Exchange返回的数据
type ExchangeDecoder func(ctx Context, resp interface{}) (statusCode int, headers http.Header, body Object, err error)
