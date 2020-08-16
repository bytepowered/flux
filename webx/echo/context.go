package echo

import (
	"github.com/bytepowered/flux/webx"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"net/url"
)

var _ webx.WebContext = new(AdaptWebContext)

type AdaptWebContext struct {
	echoc echo.Context
}

func (c *AdaptWebContext) Request() *http.Request {
	return c.echoc.Request()
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

func (c *AdaptWebContext) RequestURLPath() string {
	return c.echoc.Request().URL.Path
}

func (c *AdaptWebContext) RequestHeader() http.Header {
	return c.echoc.Request().Header
}

func (c *AdaptWebContext) RequestBody() (io.ReadCloser, error) {
	return c.echoc.Request().GetBody()
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

func (c *AdaptWebContext) FormValues() (url.Values, error) {
	return c.echoc.FormParams()
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

func (c *AdaptWebContext) Response() http.ResponseWriter {
	return c.echoc.Response()
}

func (c *AdaptWebContext) ResponseHeader() http.Header {
	return c.echoc.Response().Header()
}

func (c *AdaptWebContext) SetValue(name string, value interface{}) {
	c.echoc.Set(name, value)
}

func (c *AdaptWebContext) GetValue(name string) interface{} {
	return c.echoc.Get(name)
}

func toAdaptWebContext(echo echo.Context) webx.WebContext {
	return &AdaptWebContext{echo}
}
