package internal

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bytepowered/flux/flux-node"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"net/url"
)

var _ flux.ServerWebContext = new(AdaptWebContext)

func NewServeWebContext(ctx echo.Context, reqid string, listener flux.WebListener) flux.ServerWebContext {
	return &AdaptWebContext{
		echoc:     ctx,
		listener:  listener,
		context:   context.WithValue(ctx.Request().Context(), keyRequestId, reqid),
		variables: make(map[interface{}]interface{}, 16),
	}
}

type AdaptWebContext struct {
	listener  flux.WebListener
	context   context.Context
	echoc     echo.Context
	variables map[interface{}]interface{}
}

func (w *AdaptWebContext) WebListener() flux.WebListener {
	return w.listener
}

func (w *AdaptWebContext) ShadowContext() echo.Context {
	return w.echoc
}

func (w *AdaptWebContext) RequestId() string {
	return w.context.Value(keyRequestId).(string)
}

func (w *AdaptWebContext) Context() context.Context {
	return w.context
}

func (w *AdaptWebContext) Request() *http.Request {
	return w.echoc.Request()
}

func (w *AdaptWebContext) URI() string {
	return w.Request().RequestURI
}

func (w *AdaptWebContext) URL() *url.URL {
	return w.Request().URL
}

func (w *AdaptWebContext) Method() string {
	return w.Request().Method
}

func (w *AdaptWebContext) Host() string {
	return w.Request().Host
}

func (w *AdaptWebContext) RemoteAddr() string {
	return w.Request().RemoteAddr
}

func (w *AdaptWebContext) HeaderVars() http.Header {
	return w.Request().Header
}

func (w *AdaptWebContext) QueryVars() url.Values {
	return w.echoc.QueryParams()
}

func (w *AdaptWebContext) PathVars() url.Values {
	names := w.echoc.ParamNames()
	copied := make(url.Values, len(names))
	for _, n := range names {
		copied.Set(n, w.echoc.Param(n))
	}
	return copied
}

func (w *AdaptWebContext) FormVars() url.Values {
	f, _ := w.echoc.FormParams()
	return f
}

func (w *AdaptWebContext) CookieVars() []*http.Cookie {
	return w.echoc.Cookies()
}

func (w *AdaptWebContext) HeaderVar(name string) string {
	return w.Request().Header.Get(name)
}

func (w *AdaptWebContext) QueryVar(name string) string {
	return w.echoc.QueryParam(name)
}

func (w *AdaptWebContext) PathVar(name string) string {
	// use cached vars
	return w.echoc.Param(name)
}

func (w *AdaptWebContext) FormVar(name string) string {
	return w.echoc.FormValue(name)
}

func (w *AdaptWebContext) CookieVar(name string) (*http.Cookie, error) {
	return w.echoc.Cookie(name)
}

func (w *AdaptWebContext) BodyReader() (io.ReadCloser, error) {
	return w.Request().GetBody()
}

func (w *AdaptWebContext) Rewrite(method string, path string) {
	if "" != method {
		w.Request().Method = method
	}
	if "" != path {
		w.Request().URL.Path = path
	}
}

func (w *AdaptWebContext) Write(statusCode int, contentType string, data []byte) error {
	return w.WriteStream(statusCode, contentType, bytes.NewReader(data))
}

func (w *AdaptWebContext) WriteStream(statusCode int, contentType string, reader io.Reader) error {
	writer := w.echoc.Response()
	writer.Header().Set(echo.HeaderContentType, contentType)
	writer.WriteHeader(statusCode)
	if _, err := io.Copy(writer, reader); nil != err {
		return fmt.Errorf("web context write failed, error: %w", err)
	}
	return nil
}

func (w *AdaptWebContext) SetResponseWriter(rw http.ResponseWriter) {
	w.echoc.Response().Writer = rw
}

func (w *AdaptWebContext) ResponseWriter() http.ResponseWriter {
	return w.echoc.Response().Writer
}

func (w *AdaptWebContext) Variable(key string) interface{} {
	v, _ := w.GetVariable(key)
	return v
}

func (w *AdaptWebContext) SetVariable(key string, value interface{}) {
	w.variables[key] = value
}

func (w *AdaptWebContext) GetVariable(key string) (interface{}, bool) {
	// 本地Variable
	v, ok := w.variables[key]
	if ok {
		return v, true
	}
	// 从Context中加载
	v = w.echoc.Get(key)
	return v, nil != v
}
