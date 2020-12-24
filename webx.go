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
	ErrHttpResponseNotSupported = errors.New("webserver: http.responsewriter not supported")
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
)

// Web interfaces defines
type (
	// WebInterceptor 定义处理Web请求的中间件函数
	WebInterceptor func(WebHandler) WebHandler

	// WebHandler 定义处理Web请求的处理函数
	WebHandler func(WebContext) error

	// WebErrorHandler 定义Web服务处理异常错误的处理函数
	WebErrorHandler func(error, WebContext)

	// WebSkipper 用于部分WebInterceptor逻辑，实现忽略部分请求的功能；
	WebSkipper func(WebContext) bool

	// WebRequestBodyDecoder 解析请求体数据
	WebRequestBodyDecoder func(WebContext) url.Values
)

// WebContext 定义封装Web框架的RequestContext的接口；
// 用于 WebHandler，WebInterceptor 实现Web请求处理；
type WebContext interface {
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

	// SetRequestHeader 设置请求的Header
	SetRequestHeader(name, value string)

	// AddRequestHeader 添加请求的Header
	AddRequestHeader(name, value string)

	// RemoveRequestHeader 移除请求中的Header
	RemoveRequestHeader(name string)

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

	// Write 写入响应状态码和响应数据
	Write(statusCode int, contentType string, bytes []byte) error

	// WriteStream 写入响应状态码和流数据
	WriteStream(statusCode int, contentType string, reader io.Reader) error

	// ResponseHeader 返回响应对象的Header以及是否只读
	// 注意：部分Web框架返回只读http.Header
	ResponseHeader() (header http.Header, writable bool)

	// GetResponseHeader 获取已设置的Header键值
	GetResponseHeader(name string) string

	// SetResponseHeader 设置的Header键值
	SetResponseHeader(key, value string)

	// AddResponseHeader 添加指定Name的Header键值
	AddResponseHeader(key, value string)

	// SetResponseWriter 设置ResponseWriter
	// 如果Web框架不支持标准ResponseWriter（如fasthttp），返回 ErrHttpResponseNotSupported
	SetResponseWriter(w http.ResponseWriter) error

	// SetValue 设置Context域键值；作用域与请求生命周期相同；
	SetValue(name string, value interface{})

	// GetValue 获取Context域键值；作用域与请求生命周期相同；
	GetValue(name string) interface{}

	// HttpRequest 返回Http标准Request对象。
	// 如果Web框架不支持标准Request（如fasthttp），返回 ErrHttpRequestNotSupported
	HttpRequest() (*http.Request, error)

	// Context 返回请求的Context对象
	Context() context.Context

	// HttpResponseWriter 返回Http标准ResponseWriter对象。
	// 如果Web框架不支持标准ResponseWriter（如fasthttp），返回 ErrHttpResponseNotSupported
	HttpResponseWriter() (http.ResponseWriter, error)

	// Context 返回具体Web框架实现的WebContext对象
	RawWebContext() interface{}

	// RawWebRequest 返回具体Web框架实现的Request对象
	RawWebRequest() interface{}

	// RawWebResponse 返回具体Web框架实现的Response对象
	RawWebResponse() interface{}
}

// RawWebServer 定义Web框架服务器的接口；通过实现此接口来自定义支持不同的Web框架，用于支持不同的Web服务实现。
// 例如默认Web框架为labstack.echo；可以支持git, fasthttp等框架。
type WebServer interface {
	// SetWebErrorHandler 设置Web请求错误处理函数
	SetWebErrorHandler(h WebErrorHandler)

	// SetWebNotFoundHandler 设置Web路由不存在处理函数
	SetWebNotFoundHandler(h WebHandler)

	// SetWebRequestBodyDecoder 设置Body体解析接口
	SetWebRequestBodyDecoder(decoder WebRequestBodyDecoder)

	// AddWebInterceptor 添加全局请求拦截器，作用于路由请求前
	AddWebInterceptor(m WebInterceptor)

	// AddWebHandler 添加请求路由处理函数及其中间件
	AddWebHandler(method, pattern string, h WebHandler, m ...WebInterceptor)

	// AddWebHttpHandler 添加http标准请求路由处理函数及其中间件
	AddWebHttpHandler(method, pattern string, h http.Handler, m ...func(http.Handler) http.Handler)

	// HandleWebNotFound 处理Web无法处理路由的请求
	HandleWebNotFound(webc WebContext) error

	// RawWebServer 返回具体实现的WebServer服务对象，如echo,fasthttp的Server
	RawWebServer() interface{}

	// RawWebServer 返回具体实现的WebRouter路由处理对象，如echo,fasthttp的Router
	RawWebRouter() interface{}

	// StartTLS 启动TLS服务
	StartTLS(addr string, certFile, keyFile string) error

	// Shutdown 停止服务
	Shutdown(ctx context.Context) error
}

/// Wrapper functions

func WrapHttpHandler(h http.Handler) WebHandler {
	return func(webc WebContext) error {
		// 注意：部分Web框架不支持返回标准Request/Response
		resp, err := webc.HttpResponseWriter()
		if nil != err {
			return err
		}
		req, err := webc.HttpRequest()
		if nil != err {
			return err
		}
		h.ServeHTTP(resp, req)
		return nil
	}
}
