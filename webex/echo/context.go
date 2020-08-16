package echo

import (
	"github.com/bytepowered/flux/webex"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"net/url"
)

var _ webex.WebContext = new(AdaptEchoContext)

type AdaptEchoContext struct {
	echo echo.Context
}

func (c *AdaptEchoContext) RequestHost() string {
	return c.echo.Request().Host
}

func (c *AdaptEchoContext) RequestUserAgent() string {
	return c.echo.Request().UserAgent()
}

func (c *AdaptEchoContext) RequestRemoteAddr() string {
	return c.echo.Request().RemoteAddr
}

func (c *AdaptEchoContext) RequestHeader() http.Header {
	return c.echo.Request().Header
}

func (c *AdaptEchoContext) RequestBody() io.ReadCloser {
	return c.echo.Request().Body
}

func (c *AdaptEchoContext) RequestMethod() string {
	return c.echo.Request().Method
}

func (c *AdaptEchoContext) RequestURI() string {
	return c.echo.Request().RequestURI
}

func (c *AdaptEchoContext) RequestURL() *url.URL {
	return c.echo.Request().URL
}

func (c *AdaptEchoContext) Request() *http.Request {
	return c.echo.Request()
}

func (c *AdaptEchoContext) RemoteAddress() string {
	return c.echo.RealIP()
}

func (c *AdaptEchoContext) QueryValue(name string) string {
	return c.echo.QueryParam(name)
}

func (c *AdaptEchoContext) PathValue(name string) string {
	return c.echo.Param(name)
}

func (c *AdaptEchoContext) FormValue(name string) string {
	return c.echo.FormValue(name)
}

func (c *AdaptEchoContext) CookieValue(name string) string {
	v, err := c.echo.Cookie(name)
	if echo.ErrCookieNotFound == err {
		return ""
	} else {
		return v.Raw
	}
}

func (c *AdaptEchoContext) Response() http.ResponseWriter {
	return c.echo.Response().Writer
}

func (c *AdaptEchoContext) SetValue(name string, value interface{}) {
	c.echo.Set(name, value)
}

func (c *AdaptEchoContext) GetValue(name string) interface{} {
	return c.echo.Get(name)
}

func NewAdaptEchoContext(echo echo.Context) webex.WebContext {
	return &AdaptEchoContext{echo}
}
