package echoserver

import (
	"context"
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/internal"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"net/url"
)

const (
	__interContextKeyWebContext = "__server.core.adapted.context#890b1fa9-93ad-4b44-af24-85bcbfe646b4"
)

var _ flux.WebExchange = new(EchoWebExchange)

func NewAdaptWebExchange(id string, echoc echo.Context, server flux.WebListener, resolver flux.WebRequestBodyResolver) *EchoWebExchange {
	return &EchoWebExchange{
		context:         context.WithValue(echoc.Request().Context(), internal.ContextKeyRequestId, id),
		echoc:           echoc,
		server:          server,
		requestResolver: resolver,
	}
}

// EchoWebExchange 默认实现的基于echo框架的WebContext
// 注意：保持 EchoWebExchange 的公共访问性
type EchoWebExchange struct {
	context         context.Context
	echoc           echo.Context
	server          flux.WebListener
	requestResolver flux.WebRequestBodyResolver
	responseWriter  flux.WebResponseWriter
	pathValues      url.Values
	bodyValues      url.Values
}

func (c *EchoWebExchange) Context() context.Context {
	return c.context
}

func (c *EchoWebExchange) Method() string {
	return c.echoc.Request().Method
}

func (c *EchoWebExchange) Host() string {
	return c.echoc.Request().Host
}

func (c *EchoWebExchange) UserAgent() string {
	return c.echoc.Request().UserAgent()
}

func (c *EchoWebExchange) URI() string {
	return c.echoc.Request().RequestURI
}

func (c *EchoWebExchange) URL() *url.URL {
	return c.echoc.Request().URL
}

func (c *EchoWebExchange) Address() string {
	return c.echoc.RealIP()
}

func (c *EchoWebExchange) HeaderVars() http.Header {
	return c.echoc.Request().Header
}

func (c *EchoWebExchange) QueryVars() url.Values {
	return c.echoc.QueryParams()
}

func (c *EchoWebExchange) PathVars() url.Values {
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

func (c *EchoWebExchange) FormVars() url.Values {
	if c.bodyValues == nil {
		c.bodyValues = c.requestResolver(c)
	}
	return c.bodyValues
}

func (c *EchoWebExchange) CookieVars() []*http.Cookie {
	return c.echoc.Request().Cookies()
}

func (c *EchoWebExchange) HeaderVar(name string) string {
	return c.echoc.Request().Header.Get(name)
}

func (c *EchoWebExchange) QueryVar(name string) string {
	return c.echoc.QueryParam(name)
}

func (c *EchoWebExchange) PathVar(name string) string {
	return c.echoc.Param(name)
}

func (c *EchoWebExchange) FormVar(name string) string {
	return c.FormVars().Get(name)
}

func (c *EchoWebExchange) CookieVar(name string) *http.Cookie {
	cookie, err := c.echoc.Cookie(name)
	if err == echo.ErrCookieNotFound {
		return nil
	}
	return cookie
}

func (c *EchoWebExchange) BodyReader() (io.ReadCloser, error) {
	return c.echoc.Request().GetBody()
}

func (c *EchoWebExchange) Rewrite(method string, path string) {
	if "" != method {
		c.echoc.Request().Method = method
	}
	if "" != path {
		c.echoc.Request().URL.Path = path
	}
}

func (c *EchoWebExchange) Write(statusCode int, contentType string, bytes []byte) error {
	return c.echoc.Blob(statusCode, contentType, bytes)
}

func (c *EchoWebExchange) WriteStream(statusCode int, contentType string, reader io.Reader) error {
	return c.echoc.Stream(statusCode, contentType, reader)
}

func (c *EchoWebExchange) Send(webex flux.WebExchange, header http.Header, status int, data interface{}) error {
	return c.server.Write(webex, header, status, data)
}

func (c *EchoWebExchange) SendError(error *flux.ServeError) {
	c.server.WriteError(c, error)
}

func (c *EchoWebExchange) SetResponseHeader(key, value string) {
	c.echoc.Response().Header().Set(key, value)
}

func (c *EchoWebExchange) AddResponseHeader(key, value string) {
	c.echoc.Response().Header().Add(key, value)
}

func (c *EchoWebExchange) SetHttpResponseWriter(w http.ResponseWriter) error {
	c.echoc.Response().Writer = w
	return nil
}

func (c *EchoWebExchange) HttpResponseWriter() (http.ResponseWriter, error) {
	return c.echoc.Response().Writer, nil
}

func (c *EchoWebExchange) SetVariable(key string, value interface{}) {
	c.echoc.Set(key, value)
}

func (c *EchoWebExchange) Variable(key string) interface{} {
	return c.echoc.Get(key)
}

func (c *EchoWebExchange) RequestId() string {
	return c.context.Value(internal.ContextKeyRequestId).(string)
}

func (c *EchoWebExchange) HttpRequest() (*http.Request, error) {
	return c.echoc.Request(), nil
}

func (c *EchoWebExchange) ShadowContext() interface{} {
	return c.echoc
}

func (c *EchoWebExchange) ShadowRequest() interface{} {
	return c.echoc.Request()
}

func (c *EchoWebExchange) ShadowResponse() interface{} {
	return c.echoc.Response()
}
