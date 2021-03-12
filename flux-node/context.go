package flux

import (
	"context"
	"fmt"
	"github.com/spf13/cast"
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
)

// ServeError 定义网关处理请求的服务错误；
// 它包含：错误定义的状态码、错误消息、内部错误等元数据
type ServeError struct {
	StatusCode int                    // 响应状态码
	ErrorCode  interface{}            // 业务错误码
	Message    string                 // 错误消息
	CauseError error                  // 内部错误对象；错误对象不会被输出到请求端；
	Header     http.Header            // 响应Header
	Extras     map[string]interface{} // 用于定义和跟踪的额外信息；额外信息不会被输出到请求端；
}

func (e *ServeError) Error() string {
	if nil != e.CauseError {
		return fmt.Sprintf("ServeError: StatusCode=%d, ErrorCode=%s, Message=%s, Extras=%+v, Error=%s", e.StatusCode, e.ErrorCode, e.Message, e.Extras, e.CauseError)
	} else {
		return fmt.Sprintf("ServeError: StatusCode=%d, ErrorCode=%s, Message=%s, Extras=%+v", e.StatusCode, e.ErrorCode, e.Message, e.Extras)
	}
}

func (e *ServeError) GetErrorCode() string {
	return cast.ToString(e.ErrorCode)
}

func (e *ServeError) ExtraByKey(key string) interface{} {
	return e.Extras[key]
}

func (e *ServeError) SetExtra(key string, value interface{}) {
	if e.Extras == nil {
		e.Extras = make(map[string]interface{}, 4)
	}
	e.Extras[key] = value
}

func (e *ServeError) Merge(header http.Header) *ServeError {
	if e.Header == nil {
		e.Header = header.Clone()
	} else {
		for key, values := range header {
			for _, value := range values {
				e.Header.Add(key, value)
			}
		}
	}
	return e
}

// Request 定义请求参数读取接口
type Request interface {

	// Context 返回请求的Context对象
	Context() context.Context

	// Method 返回请求的HttpMethod
	Method() string

	// Host 返回请求的Host
	Host() string

	// UserAgent 返回请求的UserAgent
	UserAgent() string

	// URI 返回请求的URI
	URI() string

	// URL 返回请求对象的URL
	// 注意：部分Web框架返回只读url.URL
	URL() *url.URL

	// Address 返回请求对象的地址
	Address() string

	// HeaderVars 返回请求对象的Header只读；
	HeaderVars() http.Header

	// QueryVars 返回Query查询参数键值对；只读；
	QueryVars() url.Values

	// PathVars 返回动态路径参数的键值对；只读；
	PathVars() url.Values

	// FormVars 返回Form表单参数键值对；只读；
	FormVars() url.Values

	// CookieVars 返回Cookie列表；只读；
	CookieVars() []*http.Cookie

	// HeaderVar 读取请求的Header
	HeaderVar(name string) string

	// QueryVar 查询指定Name的Query参数值
	QueryVar(name string) string

	// PathValue 查询指定Name的动态路径参数值
	PathVar(name string) string

	// FormValue 查询指定Name的表单参数值
	FormVar(name string) string

	// CookieValue 查询指定Name的Cookie对象
	CookieVar(name string) *http.Cookie

	// BodyReader 返回可重复读取的Reader接口；
	BodyReader() (io.ReadCloser, error)
}

// Response 是写入响应数据的接口
type Response interface {
	// SetStatusCode 设置Http响应状态码
	SetStatusCode(status int)

	// StatusCode 获取Http响应状态码
	StatusCode() int

	// Headers 获取设置的Headers。
	HeaderVars() http.Header

	// AddHeader 添加Header键值
	AddHeader(name, value string)

	// SetHeader 设置Header键值
	SetHeader(name, value string)

	// SetPayload 设置数据响应体
	// 默认支持Reader/ReadCloser/String/Bytes类型，其它类型则序列化为JSON格式；
	SetPayload(payload interface{})

	// Body 响应数据体
	Payload() interface{}
}

// Context 定义每个请求的上下文环境
type Context interface {

	// Method 返回当前请求的Method
	Method() string

	// RequestURI 返回当前请求的URI
	URI() string

	// RequestId 返回当前请求的唯一ID
	RequestId() string

	// Request 返回请求数据接口
	Request() Request

	// Response 返回响应数据接口
	Response() Response

	// Application 返回当前Endpoint对应的应用名
	Application() string

	// Endpoint 返回当前请求路由定义的Endpoint元数据
	Endpoint() Endpoint

	// BackendService 返回BackendService信息
	BackendService() BackendService

	// BackendService 返回Endpoint Service的服务标识
	BackendServiceId() string

	// Attributes 返回所有Attributes键值对；只读；
	Attributes() map[string]interface{}

	// Attribute 获取指定key的Attribute。如果不存在，返回默认值；
	Attribute(key string, defval interface{}) interface{}

	// GetAttribute 获取指定key的Attribute，返回值和是否存在标识
	GetAttribute(key string) (interface{}, bool)

	// SetAttribute 向Context添加Attribute键值对
	SetAttribute(key string, value interface{})

	// Variable 获取指定Key的Variable。
	// 首先查找Context通过SetVariable的键值；如果不存在，则尝试查找WebContext的键值
	// 如果不存在，返回默认值；
	Variable(key string, defval interface{}) interface{}

	// GetVariable 获取当前请求范围的值；
	// 首先查找Context通过SetVariable的键值；如果不存在，则尝试查找WebContext的键值
	GetVariable(key string) (interface{}, bool)

	// SetVariable 设置当前请求范围的KV
	SetVariable(key string, value interface{})

	// Context 返回Http请求的Context对象。用于判定Http请求是否被Cancel。
	Context() context.Context

	// StartAt 返回Http请求起始的服务器时间
	StartAt() time.Time

	// AddMetric 添加路由耗时统计节点
	AddMetric(name string, elapsed time.Duration)

	// LoadMetrics 返回请求路由的的统计数据
	Metrics() []Metric

	// GetLogger 添加Context范围的Logger。
	// 通常是将关联一些追踪字段的Logger设置为ContextLogger
	SetLogger(logger Logger)

	// GetLogger 返回Context范围的Logger。
	Logger() Logger
}

// Metrics 请求路由的的统计数据
type Metric struct {
	Name    string        `json:"name"`
	Elapsed time.Duration `json:"elapsed"`
	Elapses string        `json:"elapses"`
}
