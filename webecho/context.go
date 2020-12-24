package webecho

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
	keyWebBodyDecoder = "$internal.web.adapted.body.decoder"
)

var _ flux.WebContext = new(AdaptWebContext)

func NewAdaptWebContext(echoc echo.Context, decoder flux.WebRequestBodyDecoder) *AdaptWebContext {
	echoc.Set(keyWebBodyDecoder, decoder)
	return &AdaptWebContext{
		echoc:   echoc,
		decoder: decoder,
	}
}

// AdaptWebContext 默认实现的基于echo框架的WebContext
// 注意：保持AdaptWebContext的公共访问性
type AdaptWebContext struct {
	echoc      echo.Context
	decoder    flux.WebRequestBodyDecoder
	pathValues url.Values
	bodyValues url.Values
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

func (c *AdaptWebContext) RequestURI() string {
	return c.echoc.Request().RequestURI
}

func (c *AdaptWebContext) RequestURL() (*url.URL, bool) {
	return c.echoc.Request().URL, true
}

func (c *AdaptWebContext) RequestBodyReader() (io.ReadCloser, error) {
	return c.echoc.Request().GetBody()
}

func (c *AdaptWebContext) RequestRewrite(method string, path string) {
	c.echoc.Request().Method = method
	c.echoc.Request().URL.Path = path
}

func (c *AdaptWebContext) SetRequestHeader(name, value string) {
	c.echoc.Request().Header.Set(name, value)
}

func (c *AdaptWebContext) AddRequestHeader(name, value string) {
	c.echoc.Request().Header.Add(name, value)
}

func (c *AdaptWebContext) RemoveRequestHeader(name string) {
	c.echoc.Request().Header.Del(name)
}

func (c *AdaptWebContext) HeaderValues() (http.Header, bool) {
	return c.echoc.Request().Header, true
}

func (c *AdaptWebContext) HeaderValue(name string) string {
	return c.echoc.Request().Header.Get(name)
}

func (c *AdaptWebContext) QueryValues() url.Values {
	return c.echoc.QueryParams()
}

func (c *AdaptWebContext) PathValues() url.Values {
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

func (c *AdaptWebContext) FormValues() url.Values {
	if c.bodyValues == nil {
		c.bodyValues = c.decoder(c)
	}
	return c.bodyValues
}

func (c *AdaptWebContext) CookieValues() []*http.Cookie {
	return c.echoc.Cookies()
}

func (c *AdaptWebContext) QueryValue(name string) string {
	return c.echoc.QueryParam(name)
}

func (c *AdaptWebContext) PathValue(name string) string {
	return c.echoc.Param(name)
}

func (c *AdaptWebContext) FormValue(name string) string {
	return c.FormValues().Get(name)
}

func (c *AdaptWebContext) CookieValue(name string) (*http.Cookie, bool) {
	cookie, err := c.echoc.Cookie(name)
	if err == echo.ErrCookieNotFound {
		return nil, false
	}
	return cookie, true
}

func (c *AdaptWebContext) ResponseHeader() (http.Header, bool) {
	return c.echoc.Response().Header(), true
}

func (c *AdaptWebContext) GetResponseHeader(name string) string {
	return c.echoc.Response().Header().Get(name)
}

func (c *AdaptWebContext) SetResponseHeader(name, value string) {
	c.echoc.Response().Header().Set(name, value)
}

func (c *AdaptWebContext) AddResponseHeader(name, value string) {
	c.echoc.Response().Header().Add(name, value)
}

func (c *AdaptWebContext) Write(statusCode int, contentType string, bytes []byte) (err error) {
	return c.echoc.Blob(statusCode, contentType, bytes)
}

func (c *AdaptWebContext) WriteStream(statusCode int, contentType string, reader io.Reader) error {
	return c.echoc.Stream(statusCode, contentType, reader)
}

func (c *AdaptWebContext) SetResponseWriter(w http.ResponseWriter) error {
	c.echoc.Response().Writer = w
	return nil
}

func (c *AdaptWebContext) SetValue(name string, value interface{}) {
	c.echoc.Set(name, value)
}

func (c *AdaptWebContext) GetValue(name string) interface{} {
	return c.echoc.Get(name)
}

func (c *AdaptWebContext) RawWebContext() interface{} {
	return c.echoc
}

func (c *AdaptWebContext) RawWebRequest() interface{} {
	return c.echoc.Request()
}

func (c *AdaptWebContext) RawWebResponse() interface{} {
	return c.echoc.Response()
}

func (c *AdaptWebContext) Context() context.Context {
	return c.echoc.Request().Context()
}

func (c *AdaptWebContext) HttpRequest() (*http.Request, error) {
	return c.echoc.Request(), nil
}

func (c *AdaptWebContext) HttpResponseWriter() (http.ResponseWriter, error) {
	return c.echoc.Response().Writer, nil
}

func toAdaptWebContext(echo echo.Context) flux.WebContext {
	webc, ok := echo.Get(keyWebContext).(*AdaptWebContext)
	if !ok {
		decoder, ok := echo.Get(keyWebBodyDecoder).(flux.WebRequestBodyDecoder)
		if !ok {
			decoder = DefaultRequestBodyDecoder
		}
		webc = NewAdaptWebContext(echo, decoder)
		echo.Set(keyWebContext, webc)
	}
	return webc
}
