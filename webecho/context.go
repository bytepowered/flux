package webecho

import (
	"github.com/bytepowered/flux"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"net/url"
)

var _ flux.WebContext = new(AdaptWebContext)

// AdaptWebContext 默认实现的基于echo框架的WebContext
// 注意：保持AdaptWebContext的公共访问性
type AdaptWebContext struct {
	echoc echo.Context
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

func (c *AdaptWebContext) Request() (*http.Request, error) {
	return c.echoc.Request(), nil
}

func (c *AdaptWebContext) RequestURI() string {
	return c.echoc.Request().RequestURI
}

func (c *AdaptWebContext) RequestURL() (*url.URL, bool) {
	return c.echoc.Request().URL, false
}

func (c *AdaptWebContext) RequestBodyReader() (io.ReadCloser, error) {
	return c.echoc.Request().GetBody()
}

func (c *AdaptWebContext) RequestRewrite(method string, path string) {
	c.echoc.Request().Method = method
	c.echoc.Request().URL.Path = path
}

func (c *AdaptWebContext) RequestHeader() (http.Header, bool) {
	return c.echoc.Request().Header, false
}

func (c *AdaptWebContext) GetRequestHeader(name string) string {
	return c.echoc.Request().Header.Get(name)
}

func (c *AdaptWebContext) SetRequestHeader(name, value string) {
	c.echoc.Request().Header.Set(name, value)
}

func (c *AdaptWebContext) AddRequestHeader(name, value string) {
	c.echoc.Request().Header.Add(name, value)
}

func (c *AdaptWebContext) QueryValues() url.Values {
	return c.echoc.QueryParams()
}

func (c *AdaptWebContext) PathValues() url.Values {
	names := c.echoc.ParamNames()
	values := c.echoc.ParamValues()
	pairs := make(url.Values, len(names))
	for i, name := range names {
		pairs.Set(name, values[i])
	}
	return pairs
}

func (c *AdaptWebContext) FormValues() url.Values {
	form, err := c.echoc.FormParams()
	if nil != err {
		panic(err)
	}
	return form
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
	return c.echoc.FormValue(name)
}

func (c *AdaptWebContext) CookieValue(name string) (*http.Cookie, bool) {
	cookie, err := c.echoc.Cookie(name)
	if err == echo.ErrCookieNotFound {
		return nil, false
	}
	return cookie, true
}

func (c *AdaptWebContext) ResponseHeader() (http.Header, bool) {
	return c.echoc.Response().Header(), false
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

func (c *AdaptWebContext) Response() (http.ResponseWriter, error) {
	return c.echoc.Response(), nil
}

// SetResponseWriter 设置响应的ResponseWriter
func (c *AdaptWebContext) SetResponseWriter(w http.ResponseWriter) {
	c.echoc.Response().Writer = w
}

func (c *AdaptWebContext) ResponseWrite(statusCode int, contentType string, bytes []byte) (err error) {
	return c.echoc.Blob(statusCode, contentType, bytes)
}

func (c *AdaptWebContext) ResponseWriteStream(statusCode int, contentType string, reader io.Reader) error {
	return c.echoc.Stream(statusCode, contentType, reader)
}

func (c *AdaptWebContext) ResponseNoContent(statusCode int) {
	_ = c.echoc.NoContent(statusCode)
}

func (c *AdaptWebContext) ResponseRedirect(statusCode int, url string) {
	_ = c.echoc.Redirect(statusCode, url)
}

func (c *AdaptWebContext) SetValue(name string, value interface{}) {
	c.echoc.Set(name, value)
}

func (c *AdaptWebContext) GetValue(name string) interface{} {
	return c.echoc.Get(name)
}

func (c *AdaptWebContext) Context() interface{} {
	return c.echoc
}

func toAdaptWebContext(echo echo.Context) flux.WebContext {
	return &AdaptWebContext{echo}
}
