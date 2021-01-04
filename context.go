package flux

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"
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

// Request 定义请求参数读取接口
type RequestReader interface {
	// Method 返回请求的HttpMethod
	Method() string

	// Host 返回请求的Host
	Host() string

	// UserAgent 返回请求的UserAgent
	UserAgent() string

	// RequestURI 返回请求的URI
	RequestURI() string

	// RequestURL 返回请求对象的URL
	// 注意：部分Web框架返回只读url.URL
	RequestURL() (url *url.URL, writable bool)

	// RequestBodyReader 返回可重复读取的Reader接口；
	RequestBodyReader() (io.ReadCloser, error)

	// RequestRewrite 修改请求方法和路径；
	RequestRewrite(method string, path string)

	// HeaderValues 返回请求对象的Header
	// 注意：部分Web框架返回只读http.Header
	HeaderValues() (header http.Header, writable bool)

	// QueryValues 返回Query查询参数键值对；只读；
	QueryValues() url.Values

	// PathValues 返回动态路径参数的键值对；只读；
	PathValues() url.Values

	// FormValues 返回Form表单参数键值对；只读；
	FormValues() url.Values

	// QueryValues 返回Cookie列表；只读；
	CookieValues() []*http.Cookie

	// HeaderValue 读取请求的Header
	HeaderValue(name string) string

	// QueryValue 查询指定Name的Query参数值
	QueryValue(name string) string

	// PathValue 查询指定Name的动态路径参数值
	PathValue(name string) string

	// FormValue 查询指定Name的表单参数值
	FormValue(name string) string

	// CookieValue 查询指定Name的Cookie对象，并返回是否存在标识
	CookieValue(name string) (cookie *http.Cookie, ok bool)
}

// ResponseWriter 是写入响应数据的接口
type ResponseWriter interface {
	// SetStatusCode 设置Http响应状态码
	SetStatusCode(status int)

	// StatusCode 获取Http响应状态码
	StatusCode() int

	// HeaderValues 获取设置的Headers。
	HeaderValues() http.Header

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

	// Method 返回当前请求的Method
	Method() string

	// RequestURI 返回当前请求的URI
	RequestURI() string

	// RequestId 返回当前请求的唯一ID
	RequestId() string

	// Request 返回请求数据接口
	Request() RequestReader

	// Response 返回响应数据接口
	Response() ResponseWriter

	// Endpoint 返回请求路由定义的元数据
	Endpoint() Endpoint

	// Authorize 返回当前Endpoint是否需要授权
	Authorize() bool

	// ServiceInterface 返回Endpoint Service的信息
	ServiceInterface() (proto, host, interfaceName, methodName string)

	// Protocol 返回Endpoint Service的协议名称
	ServiceProto() string

	// ServiceName 返回Endpoint Service的服务标识
	ServiceName() (interfaceName, methodName string)

	// Attributes 返回所有Attributes键值对；只读；
	Attributes() map[string]interface{}

	// GetAttribute 获取指定key的Attribute，返回值和是否存在标识
	GetAttribute(key string) (interface{}, bool)

	// SetAttribute 向Context添加Attribute键值对
	SetAttribute(name string, value interface{})

	// GetAttributeString 获取指定key的Attribute，返回值转换为String
	GetAttributeString(key string, defaultValue string) string

	// GetValue 获取当前请求范围的值；
	// 首先查找Context通过SetValue的键值；如果不存在，则尝试查找WebContext的键值
	GetValue(name string) (interface{}, bool)

	// SetValue 设置当前请求范围的KV
	SetValue(name string, value interface{})

	// GetValueString 获取当前请求范围的值；数据类型转换为String
	GetValueString(name string, defaultValue string) string

	// Context 返回Http请求的Context对象。用于判定Http请求是否被Cancel。
	Context() context.Context

	// StartTime 返回Http请求起始的服务器时间
	StartTime() time.Time

	// ElapsedTime 返回Http请求起始至今的过去时间差
	ElapsedTime() time.Duration

	// AddMetric 添加路由耗时统计节点
	AddMetric(name string, elapsed time.Duration)

	// LoadMetrics 返回请求路由的的统计数据
	LoadMetrics() []Metric

	// SetContextLogger 添加Context范围的Logger。
	// 通常是将关联一些追踪字段的Logger设置为ContextLogger
	SetContextLogger(logger Logger)

	// GetContextLogger 返回Context范围的Logger。
	GetContextLogger() (Logger, bool)
}

// Metrics 请求路由的的统计数据
type Metric struct {
	Name    string        `json:"name"`
	Elapsed time.Duration `json:"elapsed"`
	Elapses string        `json:"elapses"`
}
