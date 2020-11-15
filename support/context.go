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

func (v *ValuesRequestReader) Method() string {
	return cast.ToString(v.values["method"])
}

func (v *ValuesRequestReader) Host() string {
	return cast.ToString(v.values["host"])
}

func (v *ValuesRequestReader) UserAgent() string {
	return cast.ToString(v.values["user-agent"])
}

func (v *ValuesRequestReader) RequestURI() string {
	return cast.ToString(v.values["request-uri"])
}

func (v *ValuesRequestReader) RequestURL() (rurl *url.URL, writable bool) {
	return v.values["url"].(*url.URL), false
}

func (v *ValuesRequestReader) RequestBodyReader() (io.ReadCloser, error) {
	return v.values["body"].(io.ReadCloser), nil
}

func (v *ValuesRequestReader) RequestRewrite(method string, path string) {
	// nop
}

func (v *ValuesRequestReader) HeaderValues() (header http.Header, writable bool) {
	return v.values["header-values"].(http.Header), false
}

func (v *ValuesRequestReader) QueryValues() url.Values {
	return v.values["query-values"].(url.Values)
}

func (v *ValuesRequestReader) PathValues() url.Values {
	return v.values["path-values"].(url.Values)
}

func (v *ValuesRequestReader) FormValues() url.Values {
	return v.values["form-values"].(url.Values)
}

func (v *ValuesRequestReader) CookieValues() []*http.Cookie {
	return v.values["cookie-values"].([]*http.Cookie)
}

func (v *ValuesRequestReader) HeaderValue(name string) string {
	return cast.ToString(v.values[name])
}

func (v *ValuesRequestReader) QueryValue(name string) string {
	return cast.ToString(v.values[name])
}

func (v *ValuesRequestReader) PathValue(name string) string {
	return cast.ToString(v.values[name])
}

func (v *ValuesRequestReader) FormValue(name string) string {
	return cast.ToString(v.values[name])
}

func (v *ValuesRequestReader) CookieValue(name string) (cookie *http.Cookie, ok bool) {
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
