package internal

import (
	"github.com/bytepowered/flux"
	"github.com/labstack/echo/v4"
	"sync"
	"time"
)

// Context接口实现
type FxContext struct {
	echo         echo.Context
	endpoint     *flux.Endpoint
	requestId    string
	attrValues   *sync.Map
	scopedValues *sync.Map
	response     *FxResponse
	request      *FxRequest
}

func NewFxContext() interface{} {
	return &FxContext{
		response: newResponseWriter(),
		request:  newRequestReader(),
	}
}

func (c *FxContext) RequestReader() flux.RequestReader {
	return c.request
}

func (c *FxContext) ResponseWriter() flux.ResponseWriter {
	return c.response
}

func (c *FxContext) Endpoint() flux.Endpoint {
	return *(c.endpoint)
}

func (c *FxContext) RequestMethod() string {
	return c.echo.Request().Method
}

func (c *FxContext) RequestUri() string {
	return c.echo.Request().RequestURI
}

func (c *FxContext) RequestPath() string {
	return c.echo.Request().URL.Path
}

func (c *FxContext) RequestId() string {
	return c.requestId
}

func (c *FxContext) RequestHost() string {
	return c.echo.Request().Host
}

func (c *FxContext) AttrValues() flux.StringMap {
	m := make(flux.StringMap)
	c.attrValues.Range(func(key, value interface{}) bool {
		m[key.(string)] = value
		return true
	})
	return m
}

func (c *FxContext) SetAttrValue(name string, value interface{}) {
	c.attrValues.Store(name, value)
}

func (c *FxContext) AttrValue(name string) (interface{}, bool) {
	v, ok := c.attrValues.Load(name)
	return v, ok
}

func (c *FxContext) SetScopedValue(name string, value interface{}) {
	c.scopedValues.Store(name, value)
}

func (c *FxContext) ScopedValue(name string) (interface{}, bool) {
	v, ok := c.scopedValues.Load(name)
	return v, ok
}

func (c *FxContext) Reattach(requestId string, echo echo.Context, endpoint *flux.Endpoint) {
	httpRequest := echo.Request()
	c.echo = echo
	c.request.reattach(echo)
	c.endpoint = endpoint
	c.requestId = requestId
	c.attrValues = new(sync.Map)
	c.scopedValues = new(sync.Map)
	c.SetAttrValue(flux.XRequestTime, time.Now().Unix())
	c.SetAttrValue(flux.XRequestId, c.requestId)
	c.SetAttrValue(flux.XRequestHost, httpRequest.Host)
	c.SetAttrValue(flux.XRequestAgent, "flux/gateway")
}

func (c *FxContext) Release() {
	c.echo = nil
	c.endpoint = nil
	c.attrValues = nil
	c.scopedValues = nil
	c.request.reset()
	c.response.reset()
}
