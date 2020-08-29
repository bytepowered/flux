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

func (a *AdaptWebContext) Context() interface{} {
	return a.fastc
}

func (a *AdaptWebContext) Method() string {
	return string(a.fastc.Method())
}

func (a *AdaptWebContext) Host() string {
	return string(a.fastc.Host())
}

func (a *AdaptWebContext) UserAgent() string {
	return string(a.fastc.UserAgent())
}

func (a *AdaptWebContext) Request() (*http.Request, error) {
	return nil, flux.ErrHttpRequestNotSupported
}

func (a *AdaptWebContext) RequestURI() string {
	return string(a.fastc.RequestURI())
}

func (a *AdaptWebContext) RequestURL() *url.URL {
	//return a.fastc.Request.URI().FullURI()
	panic("implement me !")
}

func (a *AdaptWebContext) RequestHeader() (http.Header, bool) {
	l := a.fastc.Request.Header.Len()
	if a.reqHeader == nil || len(*a.reqHeader) != l {
		header := make(http.Header, l)
		a.fastc.Request.Header.VisitAll(func(key, value []byte) {
			header.Set(string(key), string(value))
		})
		a.reqHeader = &header
	}
	return *a.reqHeader, true
}

func (a *AdaptWebContext) GetRequestHeader(name string) string {
	return cast.ToString(a.fastc.Request.Header.Peek(name))
}

func (a *AdaptWebContext) SetRequestHeader(name, value string) {
	a.fastc.Request.Header.Set(name, value)
}

func (a *AdaptWebContext) AddRequestHeader(name, value string) {
	a.fastc.Request.Header.Add(name, value)
}

func (a *AdaptWebContext) RequestBodyReader() (io.ReadCloser, error) {
	r := bytes.NewReader(a.fastc.Request.Body())
	return ioutil.NopCloser(r), nil
}

func (a *AdaptWebContext) QueryValues() url.Values {
	args := a.fastc.QueryArgs()
	if a.queryValues == nil || args.Len() != len(*a.queryValues) {
		values := make(url.Values, args.Len())
		args.VisitAll(func(key, value []byte) {
			values.Set(string(key), string(value))
		})
		a.queryValues = &values
	}
	return *a.queryValues
}

func (a *AdaptWebContext) PathValues() url.Values {
	if a.pathValues == nil {
		values := make(url.Values, 16)
		a.fastc.VisitUserValues(func(key []byte, value interface{}) {
			values.Set(string(key), cast.ToString(values))
		})
		a.pathValues = &values
	}
	return *a.pathValues
}

func (a *AdaptWebContext) FormValues() url.Values {
	args := a.fastc.PostArgs()
	if a.formValues == nil || args.Len() != len(*a.formValues) {
		values := make(url.Values, args.Len())
		args.VisitAll(func(key, value []byte) {
			values.Set(string(key), string(value))
		})
		a.formValues = &values
	}
	return *a.formValues
}

func (a *AdaptWebContext) CookieValues() []*http.Cookie {
	if a.cookies == nil {
		cookies := a.fastc.Request.Header.Peek(flux.HeaderSetCookie)
		if nil == cookies || len(cookies) == 0 {
			a.cookies = make([]*http.Cookie, 0)
		} else {
			header := make(http.Header)
			header.Set(flux.HeaderSetCookie, string(cookies))
			a.cookies = (&http.Response{Header: header}).Cookies()
		}
	}
	return a.cookies
}

func (a *AdaptWebContext) QueryValue(name string) string {
	return cast.ToString(a.fastc.QueryArgs().Peek(name))
}

func (a *AdaptWebContext) PathValue(name string) string {
	return cast.ToString(a.fastc.UserValue(name))
}

func (a *AdaptWebContext) FormValue(name string) string {
	return cast.ToString(a.fastc.PostArgs().Peek(name))
}

func (a *AdaptWebContext) CookieValue(name string) (*http.Cookie, bool) {
	for _, cookie := range a.CookieValues() {
		if name == cookie.Name {
			return cookie, true
		}
	}
	return nil, false
}

func (a *AdaptWebContext) Response() (http.ResponseWriter, error) {
	return nil, flux.ErrHttpResponseNotSupported
}

func (a *AdaptWebContext) ResponseHeader() (http.Header, bool) {
	l := a.fastc.Request.Header.Len()
	if a.respHeader == nil || len(*a.respHeader) != l {
		header := make(http.Header, l)
		a.fastc.Response.Header.VisitAll(func(key, value []byte) {
			header.Set(string(key), string(value))
		})
		a.respHeader = &header
	}
	return *a.respHeader, true
}

func (a *AdaptWebContext) GetResponseHeader(name string) string {
	return cast.ToString(a.fastc.Response.Header.Peek(name))
}

func (a *AdaptWebContext) SetResponseHeader(name, value string) {
	a.fastc.Response.Header.Set(name, value)
}

func (a *AdaptWebContext) AddResponseHeader(name, value string) {
	a.fastc.Response.Header.Add(name, value)
}

func (a *AdaptWebContext) ResponseWrite(statusCode int, bytes []byte) error {
	a.fastc.SetStatusCode(statusCode)
	_, err := a.fastc.Write(bytes)
	return err
}

func (a *AdaptWebContext) SetValue(name string, value interface{}) {
	a.values[name] = value
}

func (a *AdaptWebContext) GetValue(name string) interface{} {
	return a.values[name]
}
