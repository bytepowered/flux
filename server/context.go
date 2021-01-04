package server

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/spf13/cast"
	"time"
)

var _ flux.Context = new(WrappedContext)

// Context接口实现
type WrappedContext struct {
	requestId      string
	webc           flux.WebContext
	endpoint       *flux.Endpoint
	attributes     map[string]interface{}
	values         map[string]interface{}
	requestReader  *WrappedRequestReader
	responseWriter *WrappedResponseWriter
	ctxLogger      flux.Logger
}

func NewContextWrapper() interface{} {
	return &WrappedContext{
		responseWriter: newResponseWrappedWriter(),
		requestReader:  newRequestWrappedReader(),
	}
}

func (c *WrappedContext) Request() flux.RequestReader {
	return c.requestReader
}

func (c *WrappedContext) Response() flux.ResponseWriter {
	return c.responseWriter
}

func (c *WrappedContext) Endpoint() flux.Endpoint {
	return *(c.endpoint)
}

func (c *WrappedContext) ServiceInterface() (proto, host, interfaceName, methodName string) {
	s := c.endpoint.Service
	return s.RpcProto, s.RemoteHost, s.Interface, s.Method
}

func (c *WrappedContext) ServiceProto() string {
	return c.endpoint.Service.RpcProto
}

func (c *WrappedContext) ServiceName() (interfaceName, methodName string) {
	return c.endpoint.Service.Interface, c.endpoint.Service.Method
}

func (c *WrappedContext) Authorize() bool {
	return c.endpoint.Authorize
}

func (c *WrappedContext) Method() string {
	return c.webc.Method()
}

func (c *WrappedContext) RequestURI() string {
	return c.webc.RequestURI()
}

func (c *WrappedContext) RequestId() string {
	return c.requestId
}

func (c *WrappedContext) Attributes() map[string]interface{} {
	copied := make(map[string]interface{}, len(c.attributes))
	for k, v := range c.attributes {
		copied[k] = v
	}
	return copied
}

func (c *WrappedContext) SetAttribute(name string, value interface{}) {
	c.attributes[name] = value
}

func (c *WrappedContext) GetAttribute(name string) (interface{}, bool) {
	v, ok := c.attributes[name]
	return v, ok
}

func (c *WrappedContext) GetAttributeString(name string, defaultValue string) string {
	v, ok := c.GetAttribute(name)
	if !ok {
		return defaultValue
	}
	return cast.ToString(v)
}

func (c *WrappedContext) SetValue(name string, value interface{}) {
	c.values[name] = value
}

func (c *WrappedContext) GetValue(name string) (interface{}, bool) {
	// first: Local values
	// then: WebContext values
	if lv, ok := c.values[name]; ok {
		return lv, true
	} else if cv := c.webc.GetValue(name); nil != cv {
		return cv, true
	} else {
		return nil, false
	}
}

func (c *WrappedContext) GetValueString(name string, defaultValue string) string {
	v, ok := c.GetValue(name)
	if !ok {
		return defaultValue
	}
	return cast.ToString(v)
}

func (c *WrappedContext) Context() context.Context {
	return c.webc.Context()
}

func (c *WrappedContext) SetContextLogger(logger flux.Logger) {
	c.ctxLogger = logger
}

func (c *WrappedContext) GetContextLogger() (flux.Logger, bool) {
	return c.ctxLogger, nil != c.ctxLogger
}

func (c *WrappedContext) Reattach(requestId string, webc flux.WebContext, endpoint *flux.Endpoint) {
	c.requestId = requestId
	c.webc = webc
	c.endpoint = endpoint
	c.attributes = make(map[string]interface{}, 8)
	c.values = make(map[string]interface{}, 8)
	c.requestReader.reattach(webc)
	// duplicated: c.responseWriter.reset()
	c.SetAttribute(flux.XRequestTime, time.Now().Unix())
	c.SetAttribute(flux.XRequestId, c.requestId)
	c.SetAttribute(flux.XRequestHost, webc.Host())
	c.SetAttribute(flux.XRequestAgent, "flux/gateway")
}

func (c *WrappedContext) Release() {
	c.requestId = ""
	c.webc = nil
	c.endpoint = nil
	c.attributes = nil
	c.values = nil
	c.requestReader.reset()
	c.responseWriter.reset()
	c.ctxLogger = nil
}
