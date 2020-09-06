package flux

import (
	"io"
	"net/http"
	"net/url"
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
	// Method 获取Method
	Method() string

	// Host 获取Host
	Host() string

	// UserAgent 获取UserAgent
	UserAgent() string

	// Request 获取Http请求对象。
	// 注意：部分Web框架不支持标准Request对象，返回 webx.ErrHttpRequestNotSupported 错误。
	Request() (*http.Request, error)

	// RequestURI() 获取Http请求的URI地址
	RequestURI() string

	// RequestURL 获取Http请求的URL。
	// 注意部分Web框架只能返回Readonly对象
	RequestURL() (url *url.URL, readonly bool)

	// RequestHeader 获取Http请求的全部Header。
	// 注意部分Web框架只能返回Readonly对象
	RequestHeader() (header http.Header, readonly bool)

	// 返回Http请求的Body可重复读取的接口
	RequestBodyReader() (io.ReadCloser, error)

	// QueryValue 获取Http请求的Query参数
	QueryValue(name string) string

	// PathValue 获取Http请求的Path路径参数
	PathValue(name string) string

	// FormValue 获取Http请求的Form表单参数
	FormValue(name string) string

	// HeaderValue 获取Http请求的Header参数
	HeaderValue(name string) string

	// CookieValue 获取Http请求的Cookie参数
	CookieValue(name string) string
}

// ResponseWriter 是写入响应数据的接口
type ResponseWriter interface {
	// SetStatusCode 设置Http响应状态码
	SetStatusCode(status int)

	// StatusCode 获取Http响应状态码
	StatusCode() int

	// Headers 获取设置的Headers。
	Headers() http.Header

	// AddHeader 添加Header键值
	AddHeader(name, value string)

	// SetHeader 设置Header键值
	SetHeader(name, value string)

	// SetHeaders 设置全部Headers
	SetHeaders(headers http.Header)

	// SetBody 设置数据响应体
	SetBody(body interface{})

	// Body 响应数据体
	Body() interface{}
}

// Context 定义每个请求的上下文环境
type Context interface {
	// Request 返回请求数据接口
	Request() RequestReader

	// Method 返回当前请求的Method
	Method() string

	// RequestURI 返回当前请求的URI
	RequestURI() string

	// RequestId 返回当前请求的唯一ID
	RequestId() string

	// Response 返回响应数据接口
	Response() ResponseWriter

	// Endpoint 返回请求路由定义的元数据
	Endpoint() Endpoint

	// EndpointProtoName 返回Endpoint的协议名称
	EndpointProtoName() string

	// Attributes 返回所有Attributes键值对；只读；
	Attributes() map[string]interface{}

	// GetAttribute 获取指定name的Attribute，返回值和是否存在标识
	GetAttribute(name string) (interface{}, bool)

	// SetAttribute 向Context添加Attribute键值对
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
