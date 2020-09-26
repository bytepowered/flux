package server

import (
	"github.com/bytepowered/flux"
	"sync"
	"time"
)

var _ flux.Context = new(WrappedContext)

// Context接口实现
type WrappedContext struct {
	requestId      string
	webc           flux.WebContext
	endpoint       *flux.Endpoint
	attachments    *sync.Map
	values         *sync.Map
	requestReader  *WrappedRequestReader
	responseWriter *WrappedResponseWriter
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

func (c *WrappedContext) EndpointProto() string {
	return c.endpoint.Protocol
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

func (c *WrappedContext) Attachments() map[string]interface{} {
	copied := make(map[string]interface{})
	c.attachments.Range(func(key, value interface{}) bool {
		copied[key.(string)] = value
		return true
	})
	return copied
}

func (c *WrappedContext) SetAttachment(name string, value interface{}) {
	c.attachments.Store(name, value)
}

func (c *WrappedContext) GetAttachment(name string) (interface{}, bool) {
	v, ok := c.attachments.Load(name)
	return v, ok
}

func (c *WrappedContext) SetValue(name string, value interface{}) {
	c.values.Store(name, value)
}

func (c *WrappedContext) GetValue(name string) (interface{}, bool) {
	v, ok := c.values.Load(name)
	return v, ok
}

func (c *WrappedContext) EndpointArguments() []flux.Argument {
	return c.endpoint.Arguments
}

func (c *WrappedContext) WebExchange() flux.WebContext {
	return c.webc
}

func (c *WrappedContext) Reattach(requestId string, webc flux.WebContext, endpoint *flux.Endpoint) {
	c.webc = webc
	c.requestReader.reattach(webc)
	c.endpoint = endpoint
	c.requestId = requestId
	c.attachments = new(sync.Map)
	c.values = new(sync.Map)
	c.SetAttachment(flux.XRequestTime, time.Now().Unix())
	c.SetAttachment(flux.XRequestId, c.requestId)
	c.SetAttachment(flux.XRequestHost, webc.Host())
	c.SetAttachment(flux.XRequestAgent, "flux/gateway")
}

func (c *WrappedContext) Release() {
	c.webc = nil
	c.endpoint = nil
	c.attachments = nil
	c.values = nil
	c.requestReader.reset()
	c.responseWriter.reset()
}
