package internal

import (
	"github.com/bytepowered/flux"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/random"
	"time"
)

// Context接口实现
type FxContext struct {
	echo         echo.Context
	completed    bool
	endpoint     *flux.Endpoint
	requestId    string
	attrValues   flux.StringMap
	scopedValues flux.StringMap
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
	return c.attrValues
}

func (c *FxContext) SetAttrValue(name string, value interface{}) {
	c.attrValues[name] = value
}

func (c *FxContext) AttrValue(name string) (interface{}, bool) {
	v, ok := c.attrValues[name]
	return v, ok
}

func (c *FxContext) SetScopedValue(name string, value interface{}) {
	c.scopedValues[name] = value
}

func (c *FxContext) ScopedValue(name string) (interface{}, bool) {
	v, ok := c.scopedValues[name]
	return v, ok
}

func (c *FxContext) Reattach(echo echo.Context, endpoint *flux.Endpoint) {
	httpRequest := echo.Request()
	c.echo = echo
	c.request.reattach(echo)
	c.endpoint = endpoint
	c.completed = false
	c.requestId = random.String(20)
	c.attrValues = make(flux.StringMap, 20)
	c.scopedValues = make(flux.StringMap, 8)
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
