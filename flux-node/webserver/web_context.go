package webserver

import (
	"context"
	"fmt"
	flux2 "github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/internal"
	"github.com/bytepowered/flux/flux-pkg"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"net/url"
)

const (
	ContextKeyWebPrefix     = "__webserver_core__"
	ContextKeyWebBindServer = ContextKeyWebPrefix + ".adapted.server"
	ContextKeyWebContext    = ContextKeyWebPrefix + ".adapted.context"
	ContextKeyWebResolver   = ContextKeyWebPrefix + ".adapted.body.resolver"
)

var _ flux2.WebExchange = new(AdaptWebExchange)

func NewAdaptWebExchange(id string, echoc echo.Context, server flux2.WebListener, resolver flux2.WebRequestResolver) *AdaptWebExchange {
	return &AdaptWebExchange{
		context:         context.WithValue(echoc.Request().Context(), internal.ContextKeyRequestId, id),
		echoc:           echoc,
		server:          server,
		requestResolver: resolver,
	}
}

// AdaptWebExchange 默认实现的基于echo框架的WebContext
// 注意：保持 AdaptWebExchange 的公共访问性
type AdaptWebExchange struct {
	context         context.Context
	echoc           echo.Context
	server          flux2.WebListener
	requestResolver flux2.WebRequestResolver
	responseWriter  flux2.WebResponseWriter
	pathValues      url.Values
	bodyValues      url.Values
}

func (c *AdaptWebExchange) Context() context.Context {
	return c.context
}

func (c *AdaptWebExchange) Method() string {
	return c.echoc.Request().Method
}

func (c *AdaptWebExchange) Host() string {
	return c.echoc.Request().Host
}

func (c *AdaptWebExchange) UserAgent() string {
	return c.echoc.Request().UserAgent()
}

func (c *AdaptWebExchange) URI() string {
	return c.echoc.Request().RequestURI
}

func (c *AdaptWebExchange) URL() *url.URL {
	return c.echoc.Request().URL
}

func (c *AdaptWebExchange) Address() string {
	return c.echoc.RealIP()
}

func (c *AdaptWebExchange) OnHeaderVars(access func(header http.Header)) {
	if nil != access {
		access(c.echoc.Request().Header)
	}
}

func (c *AdaptWebExchange) HeaderVars() http.Header {
	return c.echoc.Request().Header
}

func (c *AdaptWebExchange) QueryVars() url.Values {
	return c.echoc.QueryParams()
}

func (c *AdaptWebExchange) PathVars() url.Values {
	if c.pathValues == nil {
		names := c.echoc.ParamNames()
		values := c.echoc.ParamValues()
		c.pathValues = make(url.Values, len(names))
		for i, name := range names {
			c.pathValues.Set(name, values[i])
		}
	}
	return c.pathValues
}

func (c *AdaptWebExchange) FormVars() url.Values {
	if c.bodyValues == nil {
		c.bodyValues = c.requestResolver(c)
	}
	return c.bodyValues
}

func (c *AdaptWebExchange) CookieVars() []*http.Cookie {
	return c.echoc.Request().Cookies()
}

func (c *AdaptWebExchange) HeaderVar(name string) string {
	return c.echoc.Request().Header.Get(name)
}

func (c *AdaptWebExchange) QueryVar(name string) string {
	return c.echoc.QueryParam(name)
}

func (c *AdaptWebExchange) PathVar(name string) string {
	return c.echoc.Param(name)
}

func (c *AdaptWebExchange) FormVar(name string) string {
	return c.FormVars().Get(name)
}

func (c *AdaptWebExchange) CookieVar(name string) *http.Cookie {
	cookie, err := c.echoc.Cookie(name)
	if err == echo.ErrCookieNotFound {
		return nil
	}
	return cookie
}

func (c *AdaptWebExchange) BodyReader() (io.ReadCloser, error) {
	return c.echoc.Request().GetBody()
}

func (c *AdaptWebExchange) Rewrite(method string, path string) {
	if "" != method {
		c.echoc.Request().Method = method
	}
	if "" != path {
		c.echoc.Request().URL.Path = path
	}
}

func (c *AdaptWebExchange) Write(statusCode int, contentType string, bytes []byte) error {
	return c.echoc.Blob(statusCode, contentType, bytes)
}

func (c *AdaptWebExchange) WriteStream(statusCode int, contentType string, reader io.Reader) error {
	return c.echoc.Stream(statusCode, contentType, reader)
}

func (c *AdaptWebExchange) Send(webex flux2.WebExchange, header http.Header, status int, data interface{}) error {
	return c.server.Write(webex, header, status, data)
}

func (c *AdaptWebExchange) SendError(error *flux2.ServeError) {
	c.server.WriteError(c, error)
}

func (c *AdaptWebExchange) SetResponseHeader(key, value string) {
	c.echoc.Response().Header().Set(key, value)
}

func (c *AdaptWebExchange) AddResponseHeader(key, value string) {
	c.echoc.Response().Header().Add(key, value)
}

func (c *AdaptWebExchange) SetHttpResponseWriter(w http.ResponseWriter) error {
	c.echoc.Response().Writer = w
	return nil
}

func (c *AdaptWebExchange) HttpResponseWriter() (http.ResponseWriter, error) {
	return c.echoc.Response().Writer, nil
}

func (c *AdaptWebExchange) SetVariable(key string, value interface{}) {
	c.echoc.Set(key, value)
}

func (c *AdaptWebExchange) Variable(key string) interface{} {
	return c.echoc.Get(key)
}

func (c *AdaptWebExchange) RequestId() string {
	return c.context.Value(internal.ContextKeyRequestId).(string)
}

func (c *AdaptWebExchange) HttpRequest() (*http.Request, error) {
	return c.echoc.Request(), nil
}

func (c *AdaptWebExchange) ShadowContext() interface{} {
	return c.echoc
}

func (c *AdaptWebExchange) ShadowRequest() interface{} {
	return c.echoc.Request()
}

func (c *AdaptWebExchange) ShadowResponse() interface{} {
	return c.echoc.Response()
}

func toAdaptWebExchange(echoc echo.Context) flux2.WebExchange {
	if webex, ok := echoc.Get(ContextKeyWebContext).(*AdaptWebExchange); ok {
		return webex
	}
	resolver, ok := echoc.Get(ContextKeyWebResolver).(flux2.WebRequestResolver)
	fluxpkg.Assert(ok, fmt.Sprintf("invalid <request-resolver> in echo.context, was: %+v", echoc.Get(ContextKeyWebResolver)))
	server, ok := echoc.Get(ContextKeyWebBindServer).(flux2.WebListener)
	fluxpkg.Assert(ok, fmt.Sprintf("invalid <listen-server> in echo.context, was: %+v", echoc.Get(ContextKeyWebBindServer)))
	// 从Header中读取RequestId
	id := echoc.Request().Header.Get(echo.HeaderXRequestID)
	fluxpkg.Assert("" != id, "invalid <request-id> in echo.context.header")
	webex := NewAdaptWebExchange(id, echoc, server, resolver)
	echoc.Set(ContextKeyWebContext, webex)
	return webex
}
