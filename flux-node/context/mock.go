package context

import (
	"context"
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/logger"
	"github.com/spf13/cast"
	"io"
	"net/http"
	"net/url"
	"time"
)

var _ flux.Request = new(MockRequest)

func NewMockRequest(values map[string]interface{}) *MockRequest {
	return &MockRequest{values: values}
}

type MockRequest struct {
	values map[string]interface{}
}

func (r *MockRequest) Context() context.Context {
	return context.TODO()
}

func (r *MockRequest) Address() string {
	return cast.ToString(r.values["address"])
}

func (r *MockRequest) OnHeaderVars(access func(header http.Header)) {
	access(r.HeaderVars())
}

func (r *MockRequest) Method() string {
	return cast.ToString(r.values["method"])
}

func (r *MockRequest) Host() string {
	return cast.ToString(r.values["host"])
}

func (r *MockRequest) UserAgent() string {
	return cast.ToString(r.values["user-agent"])
}

func (r *MockRequest) URI() string {
	return cast.ToString(r.values["request-uri"])
}

func (r *MockRequest) URL() *url.URL {
	if v, ok := r.values["url"]; ok {
		return v.(*url.URL)
	} else {
		return nil
	}
}

func (r *MockRequest) BodyReader() (io.ReadCloser, error) {
	if v, ok := r.values["body"]; ok {
		return v.(io.ReadCloser), nil
	} else {
		return nil, nil
	}
}

func (r *MockRequest) Rewrite(method string, path string) {
	r.values["method"] = method
	r.values["path"] = path
}

func (r *MockRequest) HeaderVars() http.Header {
	if v, ok := r.values["header-values"]; ok {
		return v.(http.Header)
	} else {
		return nil
	}
}

func (r *MockRequest) QueryVars() url.Values {
	if v, ok := r.values["query-values"]; ok {
		return v.(url.Values)
	} else {
		return url.Values{}
	}
}

func (r *MockRequest) PathVars() url.Values {
	if v, ok := r.values["path-values"]; ok {
		return v.(url.Values)
	} else {
		return url.Values{}
	}
}

func (r *MockRequest) FormVars() url.Values {
	if v, ok := r.values["form-values"]; ok {
		return v.(url.Values)
	} else {
		return url.Values{}
	}
}

func (r *MockRequest) CookieVars() []*http.Cookie {
	if v, ok := r.values["cookie-values"]; ok {
		return v.([]*http.Cookie)
	} else {
		return nil
	}
}

func (r *MockRequest) HeaderVar(name string) string {
	return cast.ToString(r.values[name])
}

func (r *MockRequest) QueryVar(name string) string {
	return cast.ToString(r.values[name])
}

func (r *MockRequest) PathVar(name string) string {
	return cast.ToString(r.values[name])
}

func (r *MockRequest) FormVar(name string) string {
	return cast.ToString(r.values[name])
}

func (r *MockRequest) CookieVar(name string) *http.Cookie {
	return nil
}

////

var _ flux.Context = new(MockContext)

func NewMockContext(values map[string]interface{}) flux.Context {
	return &MockContext{
		time:      time.Now(),
		request:   NewMockRequest(values),
		ctxLogger: logger.SimpleLogger(),
	}
}

func NewEmptyContext() flux.Context {
	return NewMockContext(map[string]interface{}{})
}

type MockContext struct {
	time      time.Time
	request   *MockRequest
	ctxLogger flux.Logger
}

func (mc *MockContext) StartAt() time.Time {
	return mc.time
}

func (mc *MockContext) AddMetric(name string, elapsed time.Duration) {
	// nop
}

func (mc *MockContext) Metrics() []flux.Metric {
	return nil
}

func (mc *MockContext) Method() string {
	return mc.request.Method()
}

func (mc *MockContext) URI() string {
	return mc.request.URI()
}

func (mc *MockContext) RequestId() string {
	return cast.ToString(mc.request.values["request-id"])
}

func (mc *MockContext) Request() flux.Request {
	return mc.request
}

func (mc *MockContext) Response() flux.Response {
	return nil
}

func (mc *MockContext) Endpoint() flux.Endpoint {
	return flux.Endpoint{}
}

func (mc *MockContext) Application() string {
	return "mock"
}

func (mc *MockContext) BackendService() flux.BackendService {
	s, ok := mc.request.values["service"]
	if ok {
		return s.(flux.BackendService)
	} else {
		return flux.BackendService{}
	}
}

func (mc *MockContext) BackendServiceId() string {
	return mc.BackendService().ServiceID()
}

func (mc *MockContext) Attributes() map[string]interface{} {
	m := make(map[string]interface{}, len(mc.request.values))
	for k, v := range mc.request.values {
		m[k] = v
	}
	return m
}

func (mc *MockContext) Attribute(key string, defval interface{}) interface{} {
	if v, ok := mc.request.values[key]; ok {
		return v
	} else {
		return defval
	}
}

func (mc *MockContext) GetAttribute(key string) (interface{}, bool) {
	v, ok := mc.request.values[key]
	return v, ok
}

func (mc *MockContext) SetAttribute(name string, value interface{}) {
	mc.request.values[name] = value
}

func (mc *MockContext) Variable(name string, defval interface{}) interface{} {
	if v, ok := mc.request.values[name]; ok {
		return v
	} else {
		return defval
	}
}

func (mc *MockContext) GetVariable(name string) (interface{}, bool) {
	v, ok := mc.request.values[name]
	return v, ok
}

func (mc *MockContext) SetVariable(name string, value interface{}) {
	mc.request.values[name] = value
}

func (mc *MockContext) Context() context.Context {
	return context.Background()
}

func (mc *MockContext) SetLogger(logger flux.Logger) {
	mc.ctxLogger = logger
}

func (mc *MockContext) Logger() flux.Logger {
	return mc.ctxLogger
}
