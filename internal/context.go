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
type context struct {
	echo         echo.Context
	completed    bool
	endpoint     *flux.Endpoint
	requestId    string
	attrValues   flux.StringMap
	scopedValues flux.StringMap
	response     *response
	request      *request
	bodyBuffer   *bytes.Buffer
}

func newContext() interface{} {
	return &context{
		response:   newResponseWriter(),
		request:    newRequestReader(),
		bodyBuffer: new(bytes.Buffer),
	}
}

func (c *context) RequestReader() flux.RequestReader {
	return c.request
}

func (c *context) ResponseWriter() flux.ResponseWriter {
	return c.response
}

func (c *context) Endpoint() flux.Endpoint {
	return *(c.endpoint)
}

func (c *context) RequestMethod() string {
	return c.echo.Request().Method
}

func (c *context) RequestUri() string {
	return c.echo.Request().RequestURI
}

func (c *context) RequestId() string {
	return c.requestId
}

func (c *context) RequestHost() string {
	return c.echo.Request().Host
}

func (c *context) AttrValues() flux.StringMap {
	return c.attrValues
}

func (c *context) SetAttrValue(name string, value interface{}) {
	c.attrValues[name] = value
}

func (c *context) AttrValue(name string) (interface{}, bool) {
	v, ok := c.attrValues[name]
	return v, ok
}

func (c *context) SetScopedValue(name string, value interface{}) {
	c.scopedValues[name] = value
}

func (c *context) ScopedValue(name string) (interface{}, bool) {
	v, ok := c.scopedValues[name]
	return v, ok
}

func (c *context) bodyBufferGetter() (io.ReadCloser, error) {
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

func (c *context) reattach(echo echo.Context, endpoint *flux.Endpoint) {
	request := echo.Request()
	request.GetBody = c.bodyBufferGetter
	c.echo = echo
	c.request.attach(echo)
	c.endpoint = endpoint
	c.completed = false
	c.requestId = random.String(20)
	c.attrValues = make(flux.StringMap, 20)
	c.scopedValues = make(flux.StringMap, 8)
	c.SetAttrValue(flux.XRequestTime, time.Now().Unix())
	c.SetAttrValue(flux.XRequestId, c.requestId)
	c.SetAttrValue(flux.XRequestHost, request.Host)
	c.SetAttrValue(flux.XRequestAgent, "flux/gateway")
}

func (c *context) release() {
	c.echo = nil
	c.request = nil
	c.endpoint = nil
	c.attrValues = nil
	c.scopedValues = nil
	c.response.reset()
	c.bodyBuffer.Reset()
}
