package webserver

import (
	"context"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/internal"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/random"
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

var _ flux.WebContext = new(AdaptWebContext)

func NewAdaptWebContext(requestId string, echoc echo.Context, server flux.ListenServer, resolver flux.WebRequestResolver) *AdaptWebContext {
	return &AdaptWebContext{
		context:         context.WithValue(echoc.Request().Context(), internal.ContextKeyRequestId, requestId),
		echoc:           echoc,
		server:          server,
		requestResolver: resolver,
	}
}

// AdaptWebContext 默认实现的基于echo框架的WebContext
// 注意：保持 AdaptWebContext 的公共访问性
type AdaptWebContext struct {
	context         context.Context
	echoc           echo.Context
	server          flux.ListenServer
	requestResolver flux.WebRequestResolver
	responseWriter  flux.WebResponseWriter
	pathValues      url.Values
	bodyValues      url.Values
}

func (c *AdaptWebContext) Context() context.Context {
	return c.context
}

func (c *AdaptWebContext) Method() string {
	return c.echoc.Request().Method
}

func (c *AdaptWebContext) Host() string {
	return c.echoc.Request().Host
}

func (c *AdaptWebContext) UserAgent() string {
	return c.echoc.Request().UserAgent()
}

func (c *AdaptWebContext) URI() string {
	return c.echoc.Request().RequestURI
}

func (c *AdaptWebContext) URL() *url.URL {
	return c.echoc.Request().URL
}

func (c *AdaptWebContext) Address() string {
	return c.echoc.RealIP()
}

func (c *AdaptWebContext) OnHeaderVars(access func(header http.Header)) {
	if nil != access {
		access(c.echoc.Request().Header)
	}
}

func (c *AdaptWebContext) HeaderVars() http.Header {
	return c.echoc.Request().Header
}

func (c *AdaptWebContext) QueryVars() url.Values {
	return c.echoc.QueryParams()
}

func (c *AdaptWebContext) PathVars() url.Values {
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

func (c *AdaptWebContext) FormVars() url.Values {
	if c.bodyValues == nil {
		c.bodyValues = c.requestResolver(c)
	}
	return c.bodyValues
}

func (c *AdaptWebContext) CookieVars() []*http.Cookie {
	return c.echoc.Request().Cookies()
}

func (c *AdaptWebContext) HeaderVar(name string) string {
	return c.echoc.Request().Header.Get(name)
}

func (c *AdaptWebContext) QueryVar(name string) string {
	return c.echoc.QueryParam(name)
}

func (c *AdaptWebContext) PathVar(name string) string {
	return c.echoc.Param(name)
}

func (c *AdaptWebContext) FormVar(name string) string {
	return c.FormVars().Get(name)
}

func (c *AdaptWebContext) CookieVar(name string) *http.Cookie {
	cookie, err := c.echoc.Cookie(name)
	if err == echo.ErrCookieNotFound {
		return nil
	}
	return cookie
}

func (c *AdaptWebContext) BodyReader() (io.ReadCloser, error) {
	return c.echoc.Request().GetBody()
}

func (c *AdaptWebContext) Rewrite(method string, path string) {
	if "" != method {
		c.echoc.Request().Method = method
	}
	if "" != path {
		c.echoc.Request().URL.Path = path
	}
}

func (c *AdaptWebContext) Write(statusCode int, contentType string, bytes []byte) error {
	return c.echoc.Blob(statusCode, contentType, bytes)
}

func (c *AdaptWebContext) WriteStream(statusCode int, contentType string, reader io.Reader) error {
	return c.echoc.Stream(statusCode, contentType, reader)
}

func (c *AdaptWebContext) Send(webc flux.WebContext, header http.Header, status int, data interface{}) error {
	return c.server.Write(webc, header, status, data)
}

func (c *AdaptWebContext) SendError(error *flux.ServeError) {
	c.server.WriteError(c, error)
}

func (c *AdaptWebContext) SetResponseHeader(key, value string) {
	c.echoc.Response().Header().Set(key, value)
}

func (c *AdaptWebContext) AddResponseHeader(key, value string) {
	c.echoc.Response().Header().Add(key, value)
}

func (c *AdaptWebContext) SetHttpResponseWriter(w http.ResponseWriter) error {
	c.echoc.Response().Writer = w
	return nil
}

func (c *AdaptWebContext) HttpResponseWriter() (http.ResponseWriter, error) {
	return c.echoc.Response().Writer, nil
}

func (c *AdaptWebContext) SetVariable(key string, value interface{}) {
	c.echoc.Set(key, value)
}

func (c *AdaptWebContext) Variable(key string) interface{} {
	return c.echoc.Get(key)
}

func (c *AdaptWebContext) RequestId() string {
	return c.context.Value(internal.ContextKeyRequestId).(string)
}

func (c *AdaptWebContext) HttpRequest() (*http.Request, error) {
	return c.echoc.Request(), nil
}

func (c *AdaptWebContext) WebContext() interface{} {
	return c.echoc
}

func (c *AdaptWebContext) WebRequest() interface{} {
	return c.echoc.Request()
}

func (c *AdaptWebContext) WebResponse() interface{} {
	return c.echoc.Response()
}

func toAdaptWebContext(echoc echo.Context) flux.WebContext {
	if webc, ok := echoc.Get(ContextKeyWebContext).(*AdaptWebContext); ok {
		return webc
	}
	resolver, ok := echoc.Get(ContextKeyWebResolver).(flux.WebRequestResolver)
	if !ok {
		panic(fmt.Sprintf("invalid <request-resolver> in echo.context, was: %+v", echoc.Get(ContextKeyWebResolver)))
	}
	server, ok := echoc.Get(ContextKeyWebBindServer).(flux.ListenServer)
	if !ok {
		panic(fmt.Sprintf("invalid <listen-server> in echo.context, was: %+v", echoc.Get(ContextKeyWebBindServer)))
	}
	// 从Header中读取RequestId
	id := echoc.Request().Header.Get(echo.HeaderXRequestID)
	if "" == id {
		id = "autoid(empty)_" + random.String(32)
	}
	webc := NewAdaptWebContext(id, echoc, server, resolver)
	echoc.Set(ContextKeyWebContext, webc)
	return webc
}
