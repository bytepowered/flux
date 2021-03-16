package internal

import (
	"context"
	"github.com/bytepowered/flux/flux-node"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"net/url"
)

type ContextKey struct {
	key string
}

var (
	contextKeyRequestId = ContextKey{key: "__internal.context.request.id"}
	contextKeyPathVars  = ContextKey{key: "__internal.context.path.vars"}
)

var _ flux.ServerWebContext = new(webcontext)

func NewServeWebContext(id string, echoc echo.Context) flux.ServerWebContext {
	return &webcontext{
		echoc:     echoc,
		context:   context.WithValue(echoc.Request().Context(), contextKeyRequestId, id),
		variables: make(map[interface{}]interface{}, 16),
	}
}

type webcontext struct {
	context   context.Context
	echoc     echo.Context
	variables map[interface{}]interface{}
}

func (w *webcontext) RequestId() string {
	return w.context.Value(contextKeyRequestId).(string)
}

func (w *webcontext) Context() context.Context {
	return w.context
}

func (w *webcontext) Request() *http.Request {
	return w.echoc.Request()
}

func (w *webcontext) URI() string {
	return w.Request().RequestURI
}

func (w *webcontext) URL() *url.URL {
	return w.Request().URL
}

func (w *webcontext) Method() string {
	return w.Request().Method
}

func (w *webcontext) Host() string {
	return w.Request().Host
}

func (w *webcontext) RemoteAddr() string {
	return w.Request().RemoteAddr
}

func (w *webcontext) HeaderVars() http.Header {
	return w.Request().Header
}

func (w *webcontext) QueryVars() url.Values {
	return w.echoc.QueryParams()
}

func (w *webcontext) PathVars() url.Values {
	v, ok := w.variables[contextKeyPathVars]
	if ok {
		return v.(url.Values)
	}
	vars := make(url.Values, len(w.echoc.ParamNames()))
	for _, n := range w.echoc.ParamNames() {
		vars.Set(n, w.echoc.Param(n))
	}
	w.variables[contextKeyPathVars] = vars
	return vars
}

func (w *webcontext) FormVars() url.Values {
	f, _ := w.echoc.FormParams()
	return f
}

func (w *webcontext) CookieVars() []*http.Cookie {
	return w.echoc.Cookies()
}

func (w *webcontext) HeaderVar(name string) string {
	return w.Request().Header.Get(name)
}

func (w *webcontext) QueryVar(name string) string {
	return w.echoc.QueryParam(name)
}

func (w *webcontext) PathVar(name string) string {
	// use cached vars
	return w.PathVars().Get(name)
}

func (w *webcontext) FormVar(name string) string {
	return w.echoc.FormValue(name)
}

func (w *webcontext) CookieVar(name string) (*http.Cookie, error) {
	return w.echoc.Cookie(name)
}

func (w *webcontext) BodyReader() (io.ReadCloser, error) {
	return w.Request().GetBody()
}

func (w *webcontext) Rewrite(method string, path string) {
	if "" != method {
		w.Request().Method = method
	}
	if "" != path {
		w.Request().URL.Path = path
	}
}

func (w *webcontext) Write(statusCode int, contentType string, bytes []byte) error {
	return w.echoc.Blob(statusCode, contentType, bytes)
}

func (w *webcontext) WriteStream(statusCode int, contentType string, reader io.Reader) error {
	return w.echoc.Stream(statusCode, contentType, reader)
}

func (w *webcontext) SetResponseWriter(rw http.ResponseWriter) {
	w.echoc.Response().Writer = rw
}

func (w *webcontext) ResponseWriter() http.ResponseWriter {
	return w.echoc.Response().Writer
}

func (w *webcontext) Variable(key string) interface{} {
	v, _ := w.GetVariable(key)
	return v
}

func (w *webcontext) SetVariable(key string, value interface{}) {
	w.variables[key] = value
}

func (w *webcontext) GetVariable(key string) (interface{}, bool) {
	// 本地Variable
	v, ok := w.variables[key]
	if ok {
		return v, true
	}
	// 从Context中加载
	v = w.echoc.Get(key)
	return v, nil != v
}

//func _mockVarsLoader() url.Values {
//	return make(url.Values, 0)
//}
//
//func _mockCtxVarsLoader(key string) interface{} {
//	return nil
//}
//
//func MockWebExchange(id string) ServerWebContext {
//	mockQ := httptest.NewRequest("GET", "http://mocking/"+id, nil)
//	mockW := httptest.NewRecorder()
//	return NewServeWebExchange(id, mockQ, mockW, _mockVarsLoader, _mockVarsLoader, _mockVarsLoader, _mockCtxVarsLoader)
//}
