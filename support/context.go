package support

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/spf13/cast"
	"io"
	"net/http"
	"net/url"
)

var _ flux.RequestReader = new(ValuesRequestReader)

func NewValuesRequestReader(values map[string]interface{}) *ValuesRequestReader {
	return &ValuesRequestReader{values: values}
}

type ValuesRequestReader struct {
	values map[string]interface{}
}

func (r *ValuesRequestReader) Method() string {
	return cast.ToString(r.values["method"])
}

func (r *ValuesRequestReader) Host() string {
	return cast.ToString(r.values["host"])
}

func (r *ValuesRequestReader) UserAgent() string {
	return cast.ToString(r.values["user-agent"])
}

func (r *ValuesRequestReader) RequestURI() string {
	return cast.ToString(r.values["request-uri"])
}

func (r *ValuesRequestReader) RequestURL() (u *url.URL, writable bool) {
	if v, ok := r.values["url"]; ok {
		return v.(*url.URL), false
	} else {
		return nil, false
	}
}

func (r *ValuesRequestReader) RequestBodyReader() (io.ReadCloser, error) {
	if v, ok := r.values["body"]; ok {
		return v.(io.ReadCloser), nil
	} else {
		return nil, nil
	}
}

func (r *ValuesRequestReader) RequestRewrite(method string, path string) {
	r.values["method"] = method
	r.values["path"] = path
}

func (r *ValuesRequestReader) HeaderValues() (header http.Header, writable bool) {
	if v, ok := r.values["header-values"]; ok {
		return v.(http.Header), false
	} else {
		return header, false
	}
}

func (r *ValuesRequestReader) QueryValues() url.Values {
	if v, ok := r.values["query-values"]; ok {
		return v.(url.Values)
	} else {
		return url.Values{}
	}
}

func (r *ValuesRequestReader) PathValues() url.Values {
	if v, ok := r.values["path-values"]; ok {
		return v.(url.Values)
	} else {
		return url.Values{}
	}
}

func (r *ValuesRequestReader) FormValues() url.Values {
	if v, ok := r.values["form-values"]; ok {
		return v.(url.Values)
	} else {
		return url.Values{}
	}
}

func (r *ValuesRequestReader) CookieValues() []*http.Cookie {
	if v, ok := r.values["cookie-values"]; ok {
		return v.([]*http.Cookie)
	} else {
		return nil
	}
}

func (r *ValuesRequestReader) HeaderValue(name string) string {
	return cast.ToString(r.values[name])
}

func (r *ValuesRequestReader) QueryValue(name string) string {
	return cast.ToString(r.values[name])
}

func (r *ValuesRequestReader) PathValue(name string) string {
	return cast.ToString(r.values[name])
}

func (r *ValuesRequestReader) FormValue(name string) string {
	return cast.ToString(r.values[name])
}

func (r *ValuesRequestReader) CookieValue(name string) (cookie *http.Cookie, ok bool) {
	return nil, false
}

////

var _ flux.Context = new(ValuesContext)

func NewValuesContext(values map[string]interface{}) flux.Context {
	return &ValuesContext{
		request: NewValuesRequestReader(values),
	}
}

func NewEmptyContext() flux.Context {
	return NewValuesContext(map[string]interface{}{})
}

type ValuesContext struct {
	request   *ValuesRequestReader
	ctxLogger flux.Logger
}

func (v *ValuesContext) Method() string {
	return v.request.Method()
}

func (v *ValuesContext) RequestURI() string {
	return v.request.RequestURI()
}

func (v *ValuesContext) RequestId() string {
	return cast.ToString(v.request.values["request-id"])
}

func (v *ValuesContext) Request() flux.RequestReader {
	return v.request
}

func (v *ValuesContext) Response() flux.ResponseWriter {
	return nil
}

func (v *ValuesContext) Endpoint() flux.Endpoint {
	return flux.Endpoint{}
}

func (v *ValuesContext) Authorize() bool {
	return cast.ToBool(v.request.values["authorize"])
}

func (v *ValuesContext) ServiceInterface() (proto, host, interfaceName, methodName string) {
	return cast.ToString(v.request.values["service.proto"]),
		cast.ToString(v.request.values["service.host"]),
		cast.ToString(v.request.values["service.interface"]),
		cast.ToString(v.request.values["service.method"])
}

func (v *ValuesContext) ServiceProto() string {
	return cast.ToString(v.request.values["service.proto"])
}

func (v *ValuesContext) ServiceName() (interfaceName, methodName string) {
	return cast.ToString(v.request.values["service.interface"]), cast.ToString(v.request.values["service.method"])
}

func (v *ValuesContext) Attributes() map[string]interface{} {
	m := make(map[string]interface{}, len(v.request.values))
	for k, v := range v.request.values {
		m[k] = v
	}
	return m
}

func (v *ValuesContext) GetAttribute(key string) (interface{}, bool) {
	a, ok := v.request.values[key]
	return a, ok
}

func (v *ValuesContext) GetAttributeString(name string, defaultValue string) string {
	a, ok := v.request.values[name]
	if !ok {
		return defaultValue
	}
	return cast.ToString(a)
}

func (v *ValuesContext) SetAttribute(name string, value interface{}) {
	v.request.values[name] = value
}

func (v *ValuesContext) GetValue(name string) (interface{}, bool) {
	a, ok := v.request.values[name]
	return a, ok
}

func (v *ValuesContext) SetValue(name string, value interface{}) {
	v.request.values[name] = value
}

func (v *ValuesContext) GetValueString(name string, defaultValue string) string {
	a, ok := v.request.values[name]
	if !ok {
		return defaultValue
	}
	return cast.ToString(a)
}

func (v *ValuesContext) Context() context.Context {
	return context.Background()
}

func (v *ValuesContext) SetContextLogger(logger flux.Logger) {
	v.ctxLogger = logger
}

func (v *ValuesContext) GetContextLogger() (flux.Logger, bool) {
	return v.ctxLogger, v.ctxLogger != nil
}
