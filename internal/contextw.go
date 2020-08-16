package internal

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/webx"
	"sync"
	"time"
)

var _ flux.Context = new(ContextWrapper)

// Context接口实现
type ContextWrapper struct {
	webc       webx.WebContext
	endpoint   *flux.Endpoint
	requestId  string
	attributes *sync.Map
	values     *sync.Map
	response   *ResponseWrappedWriter
	request    *RequestWrappedReader
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

func (c *ContextWrapper) EndpointProtoName() string {
	return c.endpoint.Protocol
}

func (c *ContextWrapper) RequestMethod() string {
	return c.webc.Method()
}

func (c *ContextWrapper) RequestURI() string {
	return c.webc.RequestURI()
}

func (c *ContextWrapper) RequestURLPath() string {
	return c.webc.Request().URL.Path
}

func (c *ContextWrapper) RequestId() string {
	return c.requestId
}

func (c *ContextWrapper) RequestHost() string {
	return c.webc.Host()
}

func (c *ContextWrapper) Attributes() map[string]interface{} {
	copied := make(map[string]interface{})
	c.attributes.Range(func(key, value interface{}) bool {
		copied[key.(string)] = value
		return true
	})
	return copied
}

func (c *ContextWrapper) SetAttribute(name string, value interface{}) {
	c.attributes.Store(name, value)
}

func (c *ContextWrapper) GetAttribute(name string) (interface{}, bool) {
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

func (c *ContextWrapper) WebExchange() webx.WebContext {
	return c.webc
}

func (c *ContextWrapper) Reattach(requestId string, webc webx.WebContext, endpoint *flux.Endpoint) {
	c.webc = webc
	c.request.reattach(webc)
	c.endpoint = endpoint
	c.requestId = requestId
	c.attributes = new(sync.Map)
	c.values = new(sync.Map)
	c.SetAttribute(flux.XRequestTime, time.Now().Unix())
	c.SetAttribute(flux.XRequestId, c.requestId)
	c.SetAttribute(flux.XRequestHost, webc.Host())
	c.SetAttribute(flux.XRequestAgent, "flux/gateway")
}

func (c *ContextWrapper) Release() {
	c.webc = nil
	c.endpoint = nil
	c.attributes = nil
	c.values = nil
	c.request.reset()
	c.response.reset()
}
