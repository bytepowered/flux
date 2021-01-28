package context

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
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

func (v *MockContext) ElapsedTime() time.Duration {
	return time.Since(v.time)
}

func (v *MockContext) StartTime() time.Time {
	return v.time
}

func (v *MockContext) AddMetric(name string, elapsed time.Duration) {
	// nop
}

func (v *MockContext) Metrics() []flux.Metric {
	return nil
}

func (v *MockContext) Method() string {
	return v.request.Method()
}

func (v *MockContext) URI() string {
	return v.request.URI()
}

func (v *MockContext) RequestId() string {
	return cast.ToString(v.request.values["request-id"])
}

func (v *MockContext) Request() flux.Request {
	return v.request
}

func (v *MockContext) Response() flux.Response {
	return nil
}

func (v *MockContext) Endpoint() flux.Endpoint {
	return flux.Endpoint{}
}

func (v *MockContext) Authorize() bool {
	return cast.ToBool(v.request.values["authorize"])
}

func (v *MockContext) Service() flux.BackendService {
	s, ok := v.request.values["service"]
	if ok {
		return s.(flux.BackendService)
	} else {
		return flux.BackendService{}
	}
}

func (v *MockContext) ServiceInterface() (proto, host, interfaceName, methodName string) {
	return cast.ToString(v.request.values["service.proto"]),
		cast.ToString(v.request.values["service.host"]),
		cast.ToString(v.request.values["service.interface"]),
		cast.ToString(v.request.values["service.method"])
}

func (v *MockContext) ServiceProto() string {
	return cast.ToString(v.request.values["service.proto"])
}

func (v *MockContext) ServiceName() (interfaceName, methodName string) {
	return cast.ToString(v.request.values["service.interface"]), cast.ToString(v.request.values["service.method"])
}

func (v *MockContext) Attributes() map[string]interface{} {
	m := make(map[string]interface{}, len(v.request.values))
	for k, v := range v.request.values {
		m[k] = v
	}
	return m
}

func (v *MockContext) GetAttribute(key string) (interface{}, bool) {
	a, ok := v.request.values[key]
	return a, ok
}

func (v *MockContext) GetAttributeString(name string, defaultValue string) string {
	a, ok := v.request.values[name]
	if !ok {
		return defaultValue
	}
	return cast.ToString(a)
}

func (v *MockContext) SetAttribute(name string, value interface{}) {
	v.request.values[name] = value
}

func (v *MockContext) GetValue(name string) (interface{}, bool) {
	a, ok := v.request.values[name]
	return a, ok
}

func (v *MockContext) SetValue(name string, value interface{}) {
	v.request.values[name] = value
}

func (v *MockContext) GetValueString(name string, defaultValue string) string {
	a, ok := v.request.values[name]
	if !ok {
		return defaultValue
	}
	return cast.ToString(a)
}

func (v *MockContext) Context() context.Context {
	return context.Background()
}

func (v *MockContext) SetLogger(logger flux.Logger) {
	v.ctxLogger = logger
}

func (v *MockContext) Logger() flux.Logger {
	return v.ctxLogger
}
