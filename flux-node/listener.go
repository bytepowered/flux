package flux

import (
	"context"
	"github.com/bytepowered/flux/flux-node/internal"
	fluxpkg "github.com/bytepowered/flux/flux-pkg"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
)

// MIME types
const (
	charsetUTF8                    = "charset=UTF-8"
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
	WebHandler func(*WebExchange) error

	// WebInterceptor 定义处理Web请求的中间件函数
	WebInterceptor func(WebHandler) WebHandler

	// WebErrorHandler 定义Web服务处理异常错误的处理函数
	WebErrorHandler func(*WebExchange, error)

	// WebListenerFactory 构建 WebListener 的工厂函数，可通过Factory实现对不同Web框架的支持；
	WebListenerFactory func(string, *Configuration) WebListener

	// WebRequestIdentifier 查找请求ID的函数
	WebRequestIdentifier func(shadowContext interface{}) string

	// WebBodyResolver 解析请求体数据
	WebBodyResolver func(*WebExchange) url.Values

	// WebSkipper 用于部分WebInterceptor逻辑，实现忽略部分请求的功能；
	WebSkipper func(*WebExchange) bool
)

// NewWebExchange 构建 WebExchange 通过外部加载函数来加载请求参数
func NewWebExchange(id string, request *http.Request, response http.ResponseWriter,
	pathVars, formVars, queryVars func() url.Values, ctxVars func(key string) interface{}) *WebExchange {
	return &WebExchange{
		ctx:             context.WithValue(request.Context(), internal.ContextKeyRequestId, id),
		request:         request,
		response:        response,
		variables:       make(map[string]interface{}, 8),
		ctxVarsLoader:   ctxVars,
		pathVarsLoader:  pathVars,
		formVarsLoader:  formVars,
		queryVarsLoader: queryVars,
	}
}

type WebExchange struct {
	ctx             context.Context
	variables       map[string]interface{}
	request         *http.Request
	response        http.ResponseWriter
	ctxVarsLoader   func(string) interface{}
	pathVarsLoader  func() url.Values
	formVarsLoader  func() url.Values
	queryVarsLoader func() url.Values
}

func (w *WebExchange) RequestId() string {
	return w.ctx.Value(internal.ContextKeyRequestId).(string)
}

func (w *WebExchange) Context() context.Context {
	return w.ctx
}

func (w *WebExchange) Request() *http.Request {
	return w.request
}

func (w *WebExchange) URI() string {
	return w.request.RequestURI
}

func (w *WebExchange) URL() *url.URL {
	return w.request.URL
}

func (w *WebExchange) Method() string {
	return w.request.Method
}

func (w *WebExchange) Host() string {
	return w.request.Host
}

func (w *WebExchange) RemoteAddr() string {
	return w.request.RemoteAddr
}

// HeaderVars 返回请求对象的Header；只读；
func (w *WebExchange) HeaderVars() http.Header {
	return w.request.Header
}

// QueryVars 返回Query查询参数键值对；只读；
func (w *WebExchange) QueryVars() url.Values {
	return w.request.Form
}

// PathVars 返回动态路径参数的键值对；只读；
func (w *WebExchange) PathVars() url.Values {
	return w.pathVarsLoader()
}

// FormVars 返回Form表单参数键值对；只读；
func (w *WebExchange) FormVars() url.Values {
	return w.formVarsLoader()
}

// CookieVars 返回Cookie列表；只读；
func (w *WebExchange) CookieVars() []*http.Cookie {
	return w.request.Cookies()
}

// HeaderVar 读取请求的Header参数值
func (w *WebExchange) HeaderVar(name string) string {
	return w.request.Header.Get(name)
}

// QueryVar 查询指定Name的Query参数值
func (w *WebExchange) QueryVar(name string) string {
	return w.queryVarsLoader().Get(name)
}

// PathValue 查询指定Name的动态路径参数值
func (w *WebExchange) PathVar(name string) string {
	return w.pathVarsLoader().Get(name)
}

// FormValue 查询指定Name的表单参数值
func (w *WebExchange) FormVar(name string) string {
	return w.formVarsLoader().Get(name)
}

// CookieValue 查询指定Name的Cookie对象，并返回是否存在标识
func (w *WebExchange) CookieVar(name string) (*http.Cookie, error) {
	return w.request.Cookie(name)
}

// BodyReader 返回可重复读取的Reader接口；
func (w *WebExchange) BodyReader() (io.ReadCloser, error) {
	return w.request.GetBody()
}

// Rewrite 修改请求方法和路径；
func (w *WebExchange) Rewrite(method string, path string) {
	if "" != method {
		w.request.Method = method
	}
	if "" != path {
		w.request.URL.Path = path
	}
}

// Write 直接写入并返回响应状态码和响应数据到客户端
func (w *WebExchange) Write(statusCode int, contentType string, bytes []byte) error {
	w.setContentType(contentType)
	w.response.WriteHeader(statusCode)
	_, err := w.response.Write(bytes)
	return err
}

// WriteStream 直接写入并返回响应状态码和流数据到客户端
func (w *WebExchange) WriteStream(statusCode int, contentType string, reader io.Reader) error {
	w.setContentType(contentType)
	w.response.WriteHeader(statusCode)
	_, err := io.Copy(w.response, reader)
	return err
}

// SetResponseWriter 设置HttpWeb服务器的ResponseWriter
func (w *WebExchange) SetResponseWriter(rw http.ResponseWriter) {
	fluxpkg.AssertNotNil(rw, "<http.ResponseWriter> must not nil")
	w.response = rw
}

// ResponseWriter 返回HttpWeb服务器的ResponseWriter对象。
func (w *WebExchange) ResponseWriter() http.ResponseWriter {
	return w.response
}

// Variable 获取WebValue域键值；作用域与请求生命周期相同；
func (w *WebExchange) Variable(key string) interface{} {
	v, _ := w.GetVariable(key)
	return v
}

// SetVariable 设置Context域键值；作用域与请求生命周期相同；
func (w *WebExchange) SetVariable(key string, value interface{}) {
	w.variables[key] = value
}

// SetVariable 设置Context域键值；作用域与请求生命周期相同；
func (w *WebExchange) GetVariable(key string) (interface{}, bool) {
	// 本地Variable
	v, ok := w.variables[key]
	if ok {
		return v, true
	}
	// 从Context中加载
	v = w.ctxVarsLoader(key)
	return v, nil != v
}

func (w *WebExchange) setContentType(ct string) {
	header := w.response.Header()
	if header.Get(HeaderContentType) == "" {
		header.Set(HeaderContentType, ct)
	}
}

func _mockVarsLoader() url.Values {
	return make(url.Values, 0)
}

func _mockCtxVarsLoader(key string) interface{} {
	return nil
}

func MockWebExchange(id string) *WebExchange {
	mockQ := httptest.NewRequest("GET", "http://mocking/"+id, nil)
	mockW := httptest.NewRecorder()
	return NewWebExchange(id, mockQ, mockW, _mockVarsLoader, _mockVarsLoader, _mockVarsLoader, _mockCtxVarsLoader)
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

	// SetRequestBodyResolver 设置Body体解析接口
	SetRequestBodyResolver(decoder WebBodyResolver)

	// AddInterceptor 添加全局请求拦截器，作用于路由请求前
	AddInterceptor(m WebInterceptor)

	// AddHandler 添加请求路由处理函数及其中间件
	AddHandler(method, pattern string, h WebHandler, m ...WebInterceptor)

	// AddHttpHandler 添加http标准请求路由处理函数及其中间件
	AddHttpHandler(method, pattern string, h http.Handler, m ...func(http.Handler) http.Handler)

	// ShadowServer 返回具体实现的WebServer服务对象，如echo,fasthttp的Server
	ShadowServer() interface{}

	// ShadowRouter 返回具体实现的Router路由处理对象，如echo,fasthttp的Router
	ShadowRouter() interface{}
}

// EndpointSelector 用于请求处理前的动态选择Endpoint
type EndpointSelector interface {
	// Active 判定选择器是否激活
	Active(ctx *WebExchange, listenerId string) bool
	// DoSelect 根据请求返回Endpoint，以及是否有效标识
	DoSelect(ctx *WebExchange, listenerId string, multi *MultiEndpoint) (*Endpoint, bool)
}

// WrapHttpHandler Wrapper http.Handler to WebHandler
func WrapHttpHandler(h http.Handler) WebHandler {
	return func(webex *WebExchange) error {
		h.ServeHTTP(webex.ResponseWriter(), webex.Request())
		return nil
	}
}
