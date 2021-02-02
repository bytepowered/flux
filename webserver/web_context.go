package webserver

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"net/url"
)

const (
	keyWebContext     = "$internal.web.adapted.context"
	keyWebBodyDecoder = "$internal.web.adapted.body.resolver"
)

var _ flux.WebContext = new(AdaptWebContext)

func NewAdaptContext(echoc echo.Context, decoder flux.WebRequestResolver) *AdaptWebContext {
	echoc.Set(keyWebBodyDecoder, decoder)
	return &AdaptWebContext{
		echoc:           echoc,
		requestResolver: decoder,
	}
}

// AdaptWebContext 默认实现的基于echo框架的WebContext
// 注意：保持 AdaptWebContext 的公共访问性
type AdaptWebContext struct {
	echoc           echo.Context
	serverRef       flux.ListenServer
	requestResolver flux.WebRequestResolver
	responseWriter  flux.WebResponseWriter
	pathValues      url.Values
	bodyValues      url.Values
}

func (c *AdaptWebContext) Context() context.Context {
	return c.echoc.Request().Context()
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
	return c.serverRef.Write(webc, header, status, data)
}

func (c *AdaptWebContext) SendError(error *flux.ServeError) {
	c.serverRef.WriteError(c, error)
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

func (c *AdaptWebContext) SetScopeValue(key string, value interface{}) {
	c.echoc.Set(key, value)
}

func (c *AdaptWebContext) ScopeValue(key string) interface{} {
	return c.echoc.Get(key)
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

func toAdaptWebContext(echo echo.Context) flux.WebContext {
	webc, ok := echo.Get(keyWebContext).(*AdaptWebContext)
	if !ok {
		resolver, ok := echo.Get(keyWebBodyDecoder).(flux.WebRequestResolver)
		if !ok {
			resolver = DefaultRequestResolver
		}
		webc = NewAdaptContext(echo, resolver)
		echo.Set(keyWebContext, webc)
	}
	return webc
}
