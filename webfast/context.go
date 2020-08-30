package webfast

import (
	"bytes"
	"github.com/bytepowered/flux"
	"github.com/spf13/cast"
	"github.com/valyala/fasthttp"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

var _ flux.WebContext = new(AdaptWebContext)

func toAdaptWebContext(ctx *fasthttp.RequestCtx) flux.WebContext {
	return &AdaptWebContext{
		fastc:  ctx,
		values: make(map[string]interface{}, 16),
	}
}

// AdaptWebContext: Not thread-safe
type AdaptWebContext struct {
	fastc       *fasthttp.RequestCtx
	values      map[string]interface{}
	cookies     []*http.Cookie
	reqHeader   *http.Header
	respHeader  *http.Header
	queryValues *url.Values
	formValues  *url.Values
	pathValues  *url.Values
}

func (c *AdaptWebContext) Context() interface{} {
	return c.fastc
}

func (c *AdaptWebContext) Method() string {
	return string(c.fastc.Method())
}

func (c *AdaptWebContext) Host() string {
	return string(c.fastc.Host())
}

func (c *AdaptWebContext) UserAgent() string {
	return string(c.fastc.UserAgent())
}

func (c *AdaptWebContext) Request() (*http.Request, error) {
	return nil, flux.ErrHttpRequestNotSupported
}

func (c *AdaptWebContext) RequestBodyReader() (io.ReadCloser, error) {
	r := bytes.NewReader(c.fastc.Request.Body())
	return ioutil.NopCloser(r), nil
}

func (c *AdaptWebContext) RequestRewrite(method string, path string) {
	c.fastc.Request.Header.SetMethod(method)
	c.fastc.Request.URI().SetPath(path)
}

func (c *AdaptWebContext) RequestURI() string {
	return string(c.fastc.RequestURI())
}

func (c *AdaptWebContext) RequestURL() (*url.URL, bool) {
	stdurl, err := url.Parse(string(c.fastc.Request.URI().FullURI()))
	if nil != err {
		panic(err)
	}
	return stdurl, true
}

func (c *AdaptWebContext) RequestHeader() (http.Header, bool) {
	l := c.fastc.Request.Header.Len()
	if c.reqHeader == nil || len(*c.reqHeader) != l {
		header := make(http.Header, l)
		c.fastc.Request.Header.VisitAll(func(key, value []byte) {
			header.Set(string(key), string(value))
		})
		c.reqHeader = &header
	}
	return *c.reqHeader, true
}

func (c *AdaptWebContext) GetRequestHeader(name string) string {
	return cast.ToString(c.fastc.Request.Header.Peek(name))
}

func (c *AdaptWebContext) SetRequestHeader(name, value string) {
	c.fastc.Request.Header.Set(name, value)
}

func (c *AdaptWebContext) AddRequestHeader(name, value string) {
	c.fastc.Request.Header.Add(name, value)
}

func (c *AdaptWebContext) QueryValues() url.Values {
	args := c.fastc.QueryArgs()
	if c.queryValues == nil || args.Len() != len(*c.queryValues) {
		values := make(url.Values, args.Len())
		args.VisitAll(func(key, value []byte) {
			values.Set(string(key), string(value))
		})
		c.queryValues = &values
	}
	return *c.queryValues
}

func (c *AdaptWebContext) PathValues() url.Values {
	if c.pathValues == nil {
		values := make(url.Values, 16)
		c.fastc.VisitUserValues(func(key []byte, value interface{}) {
			values.Set(string(key), cast.ToString(values))
		})
		c.pathValues = &values
	}
	return *c.pathValues
}

func (c *AdaptWebContext) FormValues() url.Values {
	args := c.fastc.PostArgs()
	if c.formValues == nil || args.Len() != len(*c.formValues) {
		values := make(url.Values, args.Len())
		args.VisitAll(func(key, value []byte) {
			values.Set(string(key), string(value))
		})
		c.formValues = &values
	}
	return *c.formValues
}

func (c *AdaptWebContext) CookieValues() []*http.Cookie {
	if c.cookies == nil {
		cookies := c.fastc.Request.Header.Peek(flux.HeaderSetCookie)
		if nil == cookies || len(cookies) == 0 {
			c.cookies = make([]*http.Cookie, 0)
		} else {
			header := make(http.Header)
			header.Set(flux.HeaderSetCookie, string(cookies))
			c.cookies = (&http.Response{Header: header}).Cookies()
		}
	}
	return c.cookies
}

func (c *AdaptWebContext) QueryValue(name string) string {
	return cast.ToString(c.fastc.QueryArgs().Peek(name))
}

func (c *AdaptWebContext) PathValue(name string) string {
	return cast.ToString(c.fastc.UserValue(name))
}

func (c *AdaptWebContext) FormValue(name string) string {
	return cast.ToString(c.fastc.PostArgs().Peek(name))
}

func (c *AdaptWebContext) CookieValue(name string) (*http.Cookie, bool) {
	for _, cookie := range c.CookieValues() {
		if name == cookie.Name {
			return cookie, true
		}
	}
	return nil, false
}

func (c *AdaptWebContext) Response() (http.ResponseWriter, error) {
	return nil, flux.ErrHttpResponseNotSupported
}

func (c *AdaptWebContext) ResponseHeader() (http.Header, bool) {
	l := c.fastc.Request.Header.Len()
	if c.respHeader == nil || len(*c.respHeader) != l {
		header := make(http.Header, l)
		c.fastc.Response.Header.VisitAll(func(key, value []byte) {
			header.Set(string(key), string(value))
		})
		c.respHeader = &header
	}
	return *c.respHeader, true
}

func (c *AdaptWebContext) GetResponseHeader(name string) string {
	return cast.ToString(c.fastc.Response.Header.Peek(name))
}

func (c *AdaptWebContext) SetResponseHeader(name, value string) {
	c.fastc.Response.Header.Set(name, value)
}

func (c *AdaptWebContext) AddResponseHeader(name, value string) {
	c.fastc.Response.Header.Add(name, value)
}

func (c *AdaptWebContext) ResponseWrite(statusCode int, contentType string, bytes []byte) (err error) {
	c.fastc.SetStatusCode(statusCode)
	c.fastc.SetContentType(contentType)
	if nil != bytes && len(bytes) > 0 {
		c.fastc.SetBody(bytes)
	}
	// Fasthttp的响应Error由ErrorHandler处理
	return nil
}

func (c *AdaptWebContext) ResponseStream(statusCode int, contentType string, reader io.Reader) error {
	c.fastc.SetStatusCode(statusCode)
	c.fastc.SetContentType(contentType)
	c.fastc.SetBodyStream(reader, -1)
	// Fasthttp的响应Error由ErrorHandler处理
	return nil
}

func (c *AdaptWebContext) ResponseNoContent(statusCode int) {
	_ = c.ResponseWrite(statusCode, flux.MIMEApplicationJSONCharsetUTF8, nil)
}

func (c *AdaptWebContext) ResponseRedirect(statusCode int, url string) {
	c.fastc.Redirect(url, statusCode)
}

func (c *AdaptWebContext) SetValue(name string, value interface{}) {
	c.values[name] = value
}

func (c *AdaptWebContext) GetValue(name string) interface{} {
	return c.values[name]
}
