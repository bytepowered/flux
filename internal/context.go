package internal

import (
	"github.com/bytepowered/flux"
	"github.com/labstack/echo/v4"
	"sync"
	"time"
)

var _ flux.Context = new(ContextWrapper)

// Context接口实现
type ContextWrapper struct {
	context    echo.Context
	endpoint   *flux.Endpoint
	requestId  string
	attributes *sync.Map
	values     *sync.Map
	response   *ResponseWrapWriter
	request    *RequestWrapReader
}

func NewContextWrapper() interface{} {
	return &ContextWrapper{
		response: newResponseWriter(),
		request:  newRequestReader(),
	}
}

func (c *ContextWrapper) RequestReader() flux.RequestReader {
	return c.request
}

func (c *ContextWrapper) ResponseWriter() flux.ResponseWriter {
	return c.response
}

func (c *ContextWrapper) Endpoint() flux.Endpoint {
	return *(c.endpoint)
}

func (c *ContextWrapper) RequestMethod() string {
	return c.context.Request().Method
}

func (c *ContextWrapper) RequestUri() string {
	return c.context.Request().RequestURI
}

func (c *ContextWrapper) RequestPath() string {
	return c.context.Request().URL.Path
}

func (c *ContextWrapper) RequestId() string {
	return c.requestId
}

func (c *ContextWrapper) RequestHost() string {
	return c.context.Request().Host
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

func (c *ContextWrapper) HttpContext() echo.Context {
	return c.context
}

func (c *ContextWrapper) Reattach(requestId string, context echo.Context, endpoint *flux.Endpoint) {
	c.context = context
	c.request.reattach(context)
	c.endpoint = endpoint
	c.requestId = requestId
	c.attributes = new(sync.Map)
	c.values = new(sync.Map)
	c.SetAttribute(flux.XRequestTime, time.Now().Unix())
	c.SetAttribute(flux.XRequestId, c.requestId)
	c.SetAttribute(flux.XRequestHost, context.Request().Host)
	c.SetAttribute(flux.XRequestAgent, "flux/gateway")
}

func (c *ContextWrapper) Release() {
	c.context = nil
	c.endpoint = nil
	c.attributes = nil
	c.values = nil
	c.request.reset()
	c.response.reset()
}
