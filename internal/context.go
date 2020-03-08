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
type Context struct {
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

func NewContext() interface{} {
	return &Context{
		response:   newResponseWriter(),
		request:    newRequestReader(),
		bodyBuffer: new(bytes.Buffer),
	}
}

func (c *Context) RequestReader() flux.RequestReader {
	return c.request
}

func (c *Context) ResponseWriter() flux.ResponseWriter {
	return c.response
}

func (c *Context) Endpoint() flux.Endpoint {
	return *(c.endpoint)
}

func (c *Context) RequestMethod() string {
	return c.echo.Request().Method
}

func (c *Context) RequestUri() string {
	return c.echo.Request().RequestURI
}

func (c *Context) RequestId() string {
	return c.requestId
}

func (c *Context) RequestHost() string {
	return c.echo.Request().Host
}

func (c *Context) AttrValues() flux.StringMap {
	return c.attrValues
}

func (c *Context) SetAttrValue(name string, value interface{}) {
	c.attrValues[name] = value
}

func (c *Context) AttrValue(name string) (interface{}, bool) {
	v, ok := c.attrValues[name]
	return v, ok
}

func (c *Context) SetScopedValue(name string, value interface{}) {
	c.scopedValues[name] = value
}

func (c *Context) ScopedValue(name string) (interface{}, bool) {
	v, ok := c.scopedValues[name]
	return v, ok
}

func (c *Context) Reattach(echo echo.Context, endpoint *flux.Endpoint) {
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

func (c *Context) Release() {
	c.echo = nil
	c.request = nil
	c.endpoint = nil
	c.attrValues = nil
	c.scopedValues = nil
	c.response.reset()
	c.bodyBuffer.Reset()
}

func (c *Context) bodyBufferGetter() (io.ReadCloser, error) {
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
