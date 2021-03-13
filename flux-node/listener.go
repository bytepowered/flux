package flux

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
)

var (
	ErrHttpRequestNotSupported  = errors.New("webserver: http.request not supported")
	ErrHttpResponseNotSupported = errors.New("webserver: http.response-writer not supported")
)

const (
	charsetUTF8 = "charset=UTF-8"
)

// MIME types
const (
	MIMEApplicationJSON            = "application/json"
	MIMEApplicationJSONCharsetUTF8 = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEApplicationForm            = "application/x-www-form-urlencoded"
)

// Headers
const (
	HeaderAccept              = "Accept"
	HeaderAcceptEncoding      = "Accept-Encoding"
	HeaderAllow               = "Allow"
	HeaderAuthorization       = "Authorization"
	HeaderContentDisposition  = "Content-Disposition"
	HeaderContentEncoding     = "Content-Encoding"
	HeaderContentLength       = "Content-Length"
	HeaderContentType         = "Content-Type"
	HeaderCookie              = "Cookie"
	HeaderSetCookie           = "Set-Cookie"
	HeaderIfModifiedSince     = "If-Modified-Since"
	HeaderLastModified        = "Last-Modified"
	HeaderLocation            = "Location"
	HeaderUpgrade             = "Upgrade"
	HeaderVary                = "Vary"
	HeaderWWWAuthenticate     = "WWW-Authenticate"
	HeaderXForwardedFor       = "X-Forwarded-For"
	HeaderXForwardedProto     = "X-Forwarded-Protocol"
	HeaderXForwardedProtocol  = "X-Forwarded-Protocol"
	HeaderXForwardedSsl       = "X-Forwarded-Ssl"
	HeaderXUrlScheme          = "X-Url-Scheme"
	HeaderXHTTPMethodOverride = "X-HTTP-Method-Override"
	HeaderXRealIP             = "X-Real-IP"
	HeaderXRequestID          = "X-Request-ID"
	HeaderXRequestedWith      = "X-Requested-With"
	HeaderServer              = "Server"
	HeaderOrigin              = "Origin"

	// Access control
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"

	// Security
	HeaderStrictTransportSecurity         = "Strict-Transport-Security"
	HeaderXContentTypeOptions             = "X-Content-Type-Options"
	HeaderXXSSProtection                  = "X-XSS-Protection"
	HeaderXFrameOptions                   = "X-Frame-Options"
	HeaderContentSecurityPolicy           = "Content-Security-Policy"
	HeaderContentSecurityPolicyReportOnly = "Content-Security-Policy-Report-Only"
	HeaderXCSRFToken                      = "X-CSRF-Token"
	HeaderReferrerPolicy                  = "Referrer-Policy"

	// Ext
	HeaderXRequestId = "X-Request-Id"
)

// Common used status code
const (
	StatusOK           = http.StatusOK
	StatusBadRequest   = http.StatusBadRequest
	StatusNotFound     = http.StatusNotFound
	StatusUnauthorized = http.StatusUnauthorized
	StatusAccessDenied = http.StatusForbidden
	StatusServerError  = http.StatusInternalServerError
	StatusBadGateway   = http.StatusBadGateway
	StatusNoContent    = http.StatusNoContent
)

// Web interfaces defines
type (
	// WebHandler 定义处理Web请求的处理函数
	WebHandler func(WebExchange) error
	// WebInterceptor 定义处理Web请求的中间件函数
	WebInterceptor func(WebHandler) WebHandler

	// WebErrorHandler 定义Web服务处理异常错误的处理函数
	WebErrorHandler func(WebExchange, error)

	// WebListenerFactory 构建 WebListener 的工厂函数，可通过Factory实现对不同Web框架的支持；
	WebListenerFactory func(string, *Configuration) WebListener

	// WebRequestIdentifier 查找请求ID的函数
	WebRequestIdentifier func(shadowContext interface{}) string

	// WebRequestBodyResolver 解析请求体数据
	WebRequestBodyResolver func(WebExchange) url.Values

	// WebResponseWriter 用于写入Body响应数据到WebServer
	WebResponseWriter interface {
		// Write 序列化并写入数据响应
		Write(webex WebExchange, header http.Header, status int, body interface{}) error
		// WriteError 序列化并写入错误响应
		WriteError(webex WebExchange, header http.Header, status int, error *ServeError) error
	}

	// WebSkipper 用于部分WebInterceptor逻辑，实现忽略部分请求的功能；
	WebSkipper func(WebExchange) bool
)

// WebExchange 定义封装Web框架的RequestContext的接口；
// 用于 WebHandler，WebInterceptor 实现Web请求处理；
type WebExchange interface {

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

	// HeaderVars 返回请求对象的Header；只读；
	HeaderVars() http.Header

	// QueryVars 返回Query查询参数键值对；只读；
	QueryVars() url.Values

	// PathVars 返回动态路径参数的键值对；只读；
	PathVars() url.Values

	// FormVars 返回Form表单参数键值对；只读；
	FormVars() url.Values

	// CookieVars 返回Cookie列表；只读；
	CookieVars() []*http.Cookie

	// HeaderVar 读取请求的Header参数值
	HeaderVar(name string) string

	// QueryVar 查询指定Name的Query参数值
	QueryVar(name string) string

	// PathValue 查询指定Name的动态路径参数值
	PathVar(name string) string

	// FormValue 查询指定Name的表单参数值
	FormVar(name string) string

	// CookieValue 查询指定Name的Cookie对象，并返回是否存在标识
	CookieVar(name string) *http.Cookie

	// BodyReader 返回可重复读取的Reader接口；
	BodyReader() (io.ReadCloser, error)

	// Rewrite 修改请求方法和路径；
	Rewrite(method string, path string)

	// Send 发送响应体数据；
	// 由内部序列化接口处理数据体；
	Send(webex WebExchange, header http.Header, status int, data interface{}) error

	// SendError 写入错误状态数据；
	// 由内部序列化接口处理错误数据；
	SendError(error *ServeError)

	// Write 直接写入并返回响应状态码和响应数据到客户端
	Write(statusCode int, contentType string, bytes []byte) error

	// WriteStream 直接写入并返回响应状态码和流数据到客户端
	WriteStream(statusCode int, contentType string, reader io.Reader) error

	// SetResponseHeader 设置响应Response的Header键值
	SetResponseHeader(key, value string)

	// AddResponseHeader 添加响应Response指定Name的Header键值
	AddResponseHeader(key, values string)

	// SetHttpResponseWriter 设置HttpWeb服务器的ResponseWriter
	// 如果Web框架不支持标准ResponseWriter（如fasthttp），返回 ErrHttpResponseNotSupported
	SetHttpResponseWriter(w http.ResponseWriter) error

	// HttpResponseWriter 返回HttpWeb服务器的ResponseWriter对象。
	// 如果Web框架不支持标准ResponseWriter（如fasthttp），返回 ErrHttpResponseNotSupported
	HttpResponseWriter() (http.ResponseWriter, error)

	// Variable 获取WebValue域键值；作用域与请求生命周期相同；
	Variable(key string) interface{}

	// SetVariable 设置Context域键值；作用域与请求生命周期相同；
	SetVariable(key string, value interface{})

	// RequestId 获取请求RequestId
	RequestId() string

	// HttpRequest 返回Http标准Request对象。
	// 如果Web框架不支持标准Request（如fasthttp），返回 ErrHttpRequestNotSupported
	HttpRequest() (*http.Request, error)

	// ShadowContext 返回具体Web框架实现的WebContext对象
	ShadowContext() interface{}

	// ShadowRequest 返回具体Web框架实现的Request对象
	ShadowRequest() interface{}

	// ShadowResponse 返回具体Web框架实现的Response对象
	ShadowResponse() interface{}
}

// WebListener 定义Web框架服务器的接口；
// 通过实现此接口来自定义支持不同的Web框架，用于支持不同的Web服务实现。
// 例如默认Web框架为 labstack.echo；
// 可以支持git, fasthttp等框架。
type WebListener interface {

	// Id 返回服务器的ID
	ListenerId() string

	// Init 初始化服务
	Init(opts *Configuration) error

	// Listen 启动服务，监听端口
	Listen() error

	// Close 停止服务
	Close(ctx context.Context) error

	// SetErrorHandler 设置Web请求错误处理函数
	SetErrorHandler(h WebErrorHandler)

	// SetNotfoundHandler 设置Web路由不存在处理函数
	SetNotfoundHandler(h WebHandler)

	// SetResponseWriter 设置Web响应写入函数
	SetResponseWriter(WebResponseWriter)

	// SetRequestBodyResolver 设置Body体解析接口
	SetRequestBodyResolver(decoder WebRequestBodyResolver)

	// AddInterceptor 添加全局请求拦截器，作用于路由请求前
	AddInterceptor(m WebInterceptor)

	// AddHandler 添加请求路由处理函数及其中间件
	AddHandler(method, pattern string, h WebHandler, m ...WebInterceptor)

	// AddHttpHandler 添加http标准请求路由处理函数及其中间件
	AddHttpHandler(method, pattern string, h http.Handler, m ...func(http.Handler) http.Handler)

	// Write 处理并写入业务响应数据；如果发生错误，将尝试通过 WriteError 再次写入错误响应数据；
	Write(webex WebExchange, header http.Header, status int, data interface{}) error

	// WriteError 处理并写入错误响应数据
	WriteError(webex WebExchange, err *ServeError)

	// WriteNotfound 处理Web无法处理路由的请求；如果发生错误，将尝试通过 WriteError 再次写入错误响应数据；
	WriteNotfound(webex WebExchange) error

	// Server 返回具体实现的WebServer服务对象，如echo,fasthttp的Server
	ShadowServer() interface{}

	// ShadowRouter 返回具体实现的Router路由处理对象，如echo,fasthttp的Router
	ShadowRouter() interface{}
}

// EndpointSelector 用于请求处理前的动态选择Endpoint
type EndpointSelector interface {
	// Active 判定选择器是否激活
	Active(ctx WebExchange, listenerId string) bool
	// DoSelect 根据请求返回Endpoint，以及是否有效标识
	DoSelect(ctx WebExchange, listenerId string, multi *MultiEndpoint) (*Endpoint, bool)
}

// Wrapper functions

func WrapHttpHandler(h http.Handler) WebHandler {
	return func(webex WebExchange) error {
		// 注意：部分Web框架不支持返回标准Request/Response
		writer, err := webex.HttpResponseWriter()
		if nil != err {
			return err
		}
		req, err := webex.HttpRequest()
		if nil != err {
			return err
		}
		h.ServeHTTP(writer, req)
		return nil
	}
}
