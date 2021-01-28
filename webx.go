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

	// OnHeaderVars 访问请求Headers
	OnHeaderVars(access func(header http.Header))

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

	// Write 写入响应状态码和响应数据
	Write(statusCode int, contentType string, bytes []byte) error

	// WriteStream 写入响应状态码和流数据
	WriteStream(statusCode int, contentType string, reader io.Reader) error

	// SetResponseHeader 设置响应Response的Header键值
	SetResponseHeader(key, value string)

	// AddResponseHeader 添加响应Response指定Name的Header键值
	AddResponseHeader(key, values string)

	// SetResponseWriter 设置ResponseWriter
	// 如果Web框架不支持标准ResponseWriter（如fasthttp），返回 ErrHttpResponseNotSupported
	SetResponseWriter(w http.ResponseWriter) error

	// GetResponseWriter 返回Http标准ResponseWriter对象。
	// 如果Web框架不支持标准ResponseWriter（如fasthttp），返回 ErrHttpResponseNotSupported
	GetResponseWriter() (http.ResponseWriter, error)

	// SetValue 设置Context域键值；作用域与请求生命周期相同；
	SetValue(key string, value interface{})

	// GetValue 获取Context域键值；作用域与请求生命周期相同；
	GetValue(key string) interface{}

	// HttpRequest 返回Http标准Request对象。
	// 如果Web框架不支持标准Request（如fasthttp），返回 ErrHttpRequestNotSupported
	HttpRequest() (*http.Request, error)

	// WebContext 返回具体Web框架实现的WebContext对象
	WebContext() interface{}

	// WebRequest 返回具体Web框架实现的Request对象
	WebRequest() interface{}

	// WebResponse 返回具体Web框架实现的Response对象
	WebResponse() interface{}
}

// ListenServer 定义Web框架服务器的接口；
// 通过实现此接口来自定义支持不同的Web框架，用于支持不同的Web服务实现。
// 例如默认Web框架为 labstack.echo；
// 可以支持git, fasthttp等框架。
type ListenServer interface {
	// SetErrorHandler 设置Web请求错误处理函数
	SetErrorHandler(h WebErrorHandler)

	// SetNotfoundHandler 设置Web路由不存在处理函数
	SetNotfoundHandler(h WebHandler)

	// SetRequestBodyDecoder 设置Body体解析接口
	SetRequestBodyDecoder(decoder WebRequestBodyDecoder)

	// AddInterceptor 添加全局请求拦截器，作用于路由请求前
	AddInterceptor(m WebInterceptor)

	// AddHandler 添加请求路由处理函数及其中间件
	AddHandler(method, pattern string, h WebHandler, m ...WebInterceptor)

	// AddHttpHandler 添加http标准请求路由处理函数及其中间件
	AddHttpHandler(method, pattern string, h http.Handler, m ...func(http.Handler) http.Handler)

	// HandleNotfound 处理Web无法处理路由的请求
	HandleNotfound(webc WebContext) error

	// Server 返回具体实现的WebServer服务对象，如echo,fasthttp的Server
	Server() interface{}

	// Router 返回具体实现的WebRouter路由处理对象，如echo,fasthttp的Router
	Router() interface{}

	// Listen 启动服务，监听端口
	Listen(addr string, certFile, keyFile string) error

	// Close 停止服务
	Close(ctx context.Context) error
}

/// Wrapper functions

func WrapHttpHandler(h http.Handler) WebHandler {
	return func(webc WebContext) error {
		// 注意：部分Web框架不支持返回标准Request/Response
		resp, err := webc.GetResponseWriter()
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
