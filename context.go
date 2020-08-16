package flux

import (
	"io"
	"net/http"
	"strings"
)

const (
	XRequestId    = "X-Request-Id"
	XRequestTime  = "X-Request-Time"
	XRequestHost  = "X-Request-Host"
	XRequestAgent = "X-Request-Agent"
	XJwtSubject   = "X-Jwt-Subject"
	XJwtIssuer    = "X-Jwt-Issuer"
	XJwtToken     = "X-Jwt-Token"
)

const (
	StatusOK           = http.StatusOK
	StatusBadRequest   = http.StatusBadRequest
	StatusNotFound     = http.StatusNotFound
	StatusUnauthorized = http.StatusUnauthorized
	StatusAccessDenied = http.StatusForbidden
	StatusServerError  = http.StatusInternalServerError
	StatusBadGateway   = http.StatusBadGateway
)

// Request 定义请求参数读取接口
type RequestReader interface {
	// 获取Method
	Method() string
	// 获取Host
	Host() string
	// 获取UserAgent
	UserAgent() string
	// 获取Http请求对象
	Request() *http.Request
	// 获取Http请求的URI地址
	RequestURI() string
	// 获取Http请求的全部Header
	RequestHeader() http.Header
	// 返回Http请求的Body可重复读取的接口
	RequestBody() (io.ReadCloser, error)

	// 获取Http请求的Query参数
	QueryValue(name string) string
	// 获取Http请求的Path路径参数
	PathValue(name string) string
	// 获取Http请求的Form表单参数
	FormValue(name string) string
	// 获取Http请求的Header参数
	HeaderValue(name string) string
	// 获取Http请求的Cookie参数
	CookieValue(name string) string
}

// Response 是写入响应数据的接口
type ResponseWriter interface {
	// SetStatusCode 设置Http响应状态码
	SetStatusCode(status int) ResponseWriter
	// StatusCode 获取Http响应状态码
	StatusCode() int
	// AddHeader 添加Header
	AddHeader(name, value string) ResponseWriter
	// SetHeaders 设置全部Headers
	SetHeaders(headers http.Header) ResponseWriter
	// Header 获取设置的Headers
	Headers() http.Header
	// SetBody 设置数据响应体
	SetBody(body interface{}) ResponseWriter
	// Body 响应数据体
	Body() interface{}
}

// Context 定义每个请求的上下文环境
type Context interface {
	// 返回当前请求的Method
	RequestMethod() string
	// 返回当前请求的Host地址
	RequestHost() string
	// 返回当前请求的URI
	RequestURI() string
	// 返回当前请求的URI的路径
	RequestURLPath() string
	// 返回当前请求的唯一ID
	RequestId() string
	// 返回请求数据接口
	Request() RequestReader
	// 返回响应数据接口
	Response() ResponseWriter
	// 返回请求路由定义的元数据
	Endpoint() Endpoint
	// 返回Endpoint的协议名称
	EndpointProtoName() string
	// 返回所有Attributes键值对
	Attributes() map[string]interface{}
	// 获取指定name的Attribute，返回值和是否存在标识
	GetAttribute(name string) (interface{}, bool)
	// 向Context添加Attribute键值对
	SetAttribute(name string, value interface{})
	// 获取当前请求范围的值
	GetValue(name string) (interface{}, bool)
	// 设置当前请求范围的KV
	SetValue(name string, value interface{})
}

// LookupValue 搜索Lookup指定域的值。支持：
// 1. query:<name>
// 2. form:<name>
// 3. path:<name>
// 4. header:<name>
// 5. attr:<name>
func LookupValue(lookup string, ctx Context) interface{} {
	if "" == lookup || nil == ctx {
		return nil
	}
	req := ctx.Request()
	parts := strings.Split(lookup, ":")
	if len(parts) == 1 {
		return req.HeaderValue(parts[0])
	}
	switch strings.ToUpper(parts[0]) {
	case ScopeQuery:
		return req.QueryValue(parts[1])
	case ScopeForm:
		return req.FormValue(parts[1])
	case ScopePath:
		return req.PathValue(parts[1])
	case ScopeHeader:
		return req.HeaderValue(parts[1])
	case ScopeAttr:
		v, _ := ctx.GetAttribute(parts[1])
		return v
	default:
		return nil
	}
}
