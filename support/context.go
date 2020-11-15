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
	// nop
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
		reader: NewValuesRequestReader(values),
	}
}

func NewEmptyContext() flux.Context {
	return NewValuesContext(map[string]interface{}{})
}

type ValuesContext struct {
	reader *ValuesRequestReader
}

func (v *ValuesContext) Method() string {
	return v.reader.Method()
}

func (v *ValuesContext) RequestURI() string {
	return v.reader.RequestURI()
}

func (v *ValuesContext) RequestId() string {
	return cast.ToString(v.reader.values["request-id"])
}

func (v *ValuesContext) Request() flux.RequestReader {
	return v.reader
}

func (v *ValuesContext) Response() flux.ResponseWriter {
	panic("not supported")
}

func (v *ValuesContext) Endpoint() flux.Endpoint {
	panic("not supported")
}

func (v *ValuesContext) Authorize() bool {
	return cast.ToBool(v.reader.values["authorize"])
}

func (v *ValuesContext) ServiceInterface() (proto, host, interfaceName, methodName string) {
	return cast.ToString(v.reader.values["service.proto"]),
		cast.ToString(v.reader.values["service.host"]),
		cast.ToString(v.reader.values["service.interface"]),
		cast.ToString(v.reader.values["service.method"])
}

func (v *ValuesContext) ServiceProto() string {
	return cast.ToString(v.reader.values["service.proto"])
}

func (v *ValuesContext) ServiceName() (interfaceName, methodName string) {
	return cast.ToString(v.reader.values["service.interface"]), cast.ToString(v.reader.values["service.method"])
}

func (v *ValuesContext) Attributes() map[string]interface{} {
	m := make(map[string]interface{}, len(v.reader.values))
	for k, v := range v.reader.values {
		m[k] = v
	}
	return m
}

func (v *ValuesContext) GetAttribute(key string) (interface{}, bool) {
	a, ok := v.reader.values[key]
	return a, ok
}

func (v *ValuesContext) SetAttribute(name string, value interface{}) {
	v.reader.values[name] = value
}

func (v *ValuesContext) GetValue(name string) (interface{}, bool) {
	a, ok := v.reader.values[name]
	return a, ok
}

func (v *ValuesContext) SetValue(name string, value interface{}) {
	v.reader.values[name] = value
}

func (v *ValuesContext) Context() context.Context {
	return context.Background()
}

func (v *ValuesContext) SetContextLogger(logger flux.Logger) {
	panic("not supported")
}

func (v *ValuesContext) GetContextLogger() (flux.Logger, bool) {
	return nil, false
}
