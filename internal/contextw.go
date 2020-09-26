package internal

import (
	"github.com/bytepowered/flux"
	"sync"
	"time"
)

var _ flux.Context = new(ContextWrapper)

// Context接口实现
type ContextWrapper struct {
	webc       flux.WebContext
	endpoint   *flux.Endpoint
	requestId  string
	attributes *sync.Map
	values     *sync.Map
	response   *WrappedResponseWriter
	request    *WrappedRequestReader
}

func NewContextWrapper() interface{} {
	return &ContextWrapper{
		response: newResponseWrappedWriter(),
		request:  newRequestWrappedReader(),
	}
}

func (c *ContextWrapper) Request() flux.RequestReader {
	return c.request
}

func (c *ContextWrapper) Response() flux.ResponseWriter {
	return c.response
}

func (c *ContextWrapper) Endpoint() flux.Endpoint {
	return *(c.endpoint)
}

func (c *ContextWrapper) EndpointProto() string {
	return c.endpoint.Protocol
}

func (c *ContextWrapper) Method() string {
	return c.webc.Method()
}

func (c *ContextWrapper) RequestURI() string {
	return c.webc.RequestURI()
}

func (c *ContextWrapper) RequestId() string {
	return c.requestId
}

func (c *ContextWrapper) Attachments() map[string]interface{} {
	copied := make(map[string]interface{})
	c.attributes.Range(func(key, value interface{}) bool {
		copied[key.(string)] = value
		return true
	})
	return copied
}

func (c *ContextWrapper) SetAttachment(name string, value interface{}) {
	c.attributes.Store(name, value)
}

func (c *ContextWrapper) GetAttachment(name string) (interface{}, bool) {
	v, ok := c.attributes.Load(name)
	return v, ok
}

func (c *ContextWrapper) SetValue(name string, value interface{}) {
	c.values.Store(name, value)
}

func (c *ContextWrapper) GetValue(name string) (interface{}, bool) {
	v, ok := c.values.Load(name)
	return v, ok
}

func (c *ContextWrapper) EndpointArguments() []flux.Argument {
	return c.endpoint.Arguments
}

func (c *ContextWrapper) WebExchange() flux.WebContext {
	return c.webc
}

func (c *ContextWrapper) Reattach(requestId string, webc flux.WebContext, endpoint *flux.Endpoint) {
	c.webc = webc
	c.request.reattach(webc)
	c.endpoint = endpoint
	c.requestId = requestId
	c.attributes = new(sync.Map)
	c.values = new(sync.Map)
	c.SetAttachment(flux.XRequestTime, time.Now().Unix())
	c.SetAttachment(flux.XRequestId, c.requestId)
	c.SetAttachment(flux.XRequestHost, webc.Host())
	c.SetAttachment(flux.XRequestAgent, "flux/gateway")
}

func (c *ContextWrapper) Release() {
	c.webc = nil
	c.endpoint = nil
	c.attributes = nil
	c.values = nil
	c.request.reset()
	c.response.reset()
}
