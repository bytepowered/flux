package internal

import (
	"bytes"
	"github.com/bytepowered/flux"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/random"
	"io"
	"io/ioutil"
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
	bodyBuffer   *bytes.Buffer
}

func NewFxContext() interface{} {
	return &FxContext{
		response:   newResponseWriter(),
		request:    newRequestReader(),
		bodyBuffer: new(bytes.Buffer),
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
	httpRequest.GetBody = c.bodyBufferGetter
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
	c.bodyBuffer.Reset()
}

func (c *FxContext) bodyBufferGetter() (io.ReadCloser, error) {
	if c.bodyBuffer != nil {
		return ioutil.NopCloser(bytes.NewReader(c.bodyBuffer.Bytes())), nil
	}
	request := c.echo.Request()
	if request.Body != nil {
		b, err := ioutil.ReadAll(request.Body)
		if err != nil {
			return nil, err
		}
		// Restore the Body
		if c, ok := request.Body.(io.Closer); ok {
			_ = c.Close()
		}
		request.Body = ioutil.NopCloser(bytes.NewBuffer(b))
		return ioutil.NopCloser(bytes.NewBuffer(b)), nil
	}
	return nil, nil
}
