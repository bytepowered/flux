package server

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/spf13/cast"
	"sync"
	"time"
)

var _ flux.Context = new(DefaultContext)

// Context接口实现
type DefaultContext struct {
	requestId      string
	webc           flux.WebContext
	endpoint       *flux.Endpoint
	attributes     *sync.Map
	values         *sync.Map
	metrics        []flux.Metric
	beginTime      time.Time
	requestReader  *DefaultRequestReader
	responseWriter *DefaultResponseWriter
	ctxLogger      flux.Logger
}

func DefaultContextFactory() flux.Context {
	return &DefaultContext{
		responseWriter: NewDefaultResponseWriter(),
		requestReader:  NewDefaultRequestReader(),
	}
}

func (c *DefaultContext) Request() flux.RequestReader {
	return c.requestReader
}

func (c *DefaultContext) Response() flux.ResponseWriter {
	return c.responseWriter
}

func (c *DefaultContext) Endpoint() flux.Endpoint {
	return *(c.endpoint)
}

func (c *DefaultContext) ServiceInterface() (proto, host, interfaceName, methodName string) {
	s := c.endpoint.Service
	return s.AttrRpcProto(), s.RemoteHost, s.Interface, s.Method
}

func (c *DefaultContext) ServiceProto() string {
	return c.endpoint.Service.AttrRpcProto()
}

func (c *DefaultContext) ServiceName() (interfaceName, methodName string) {
	return c.endpoint.Service.Interface, c.endpoint.Service.Method
}

func (c *DefaultContext) Authorize() bool {
	return c.endpoint.AttrAuthorize()
}

func (c *DefaultContext) Method() string {
	return c.webc.Method()
}

func (c *DefaultContext) RequestURI() string {
	return c.webc.RequestURI()
}

func (c *DefaultContext) RequestId() string {
	return c.requestId
}

func (c *DefaultContext) Attributes() map[string]interface{} {
	copied := make(map[string]interface{}, 16)
	c.attributes.Range(func(k, v interface{}) bool {
		copied[k.(string)] = v
		return true
	})
	return copied
}

func (c *DefaultContext) SetAttribute(name string, value interface{}) {
	c.attributes.Store(name, value)
}

func (c *DefaultContext) GetAttribute(name string) (interface{}, bool) {
	v, ok := c.attributes.Load(name)
	return v, ok
}

func (c *DefaultContext) GetAttributeString(name string, defaultValue string) string {
	v, ok := c.GetAttribute(name)
	if !ok {
		return defaultValue
	}
	return cast.ToString(v)
}

func (c *DefaultContext) SetValue(name string, value interface{}) {
	c.values.Store(name, value)
}

func (c *DefaultContext) GetValue(name string) (interface{}, bool) {
	// first: Local values
	// then: WebContext values
	if lv, ok := c.values.Load(name); ok {
		return lv, true
	} else if cv := c.webc.GetValue(name); nil != cv {
		return cv, true
	} else {
		return nil, false
	}
}

func (c *DefaultContext) GetValueString(name string, defaultValue string) string {
	v, ok := c.GetValue(name)
	if !ok {
		return defaultValue
	}
	return cast.ToString(v)
}

func (c *DefaultContext) Context() context.Context {
	return c.webc.Context()
}

func (c *DefaultContext) LoadMetrics() []flux.Metric {
	dist := make([]flux.Metric, len(c.metrics))
	copy(dist, c.metrics)
	return dist
}

func (c *DefaultContext) SetContextLogger(logger flux.Logger) {
	c.ctxLogger = logger
}

func (c *DefaultContext) GetContextLogger() (flux.Logger, bool) {
	return c.ctxLogger, nil != c.ctxLogger
}

func (c *DefaultContext) StartTime() time.Time {
	return c.beginTime
}

func (c *DefaultContext) ElapsedTime() time.Duration {
	return time.Since(c.beginTime)
}

func (c *DefaultContext) AddMetric(name string, elapsed time.Duration) {
	c.metrics = append(c.metrics, flux.Metric{
		Name: name, Elapsed: elapsed, Elapses: elapsed.String(),
	})
}

func (c *DefaultContext) Reattach(requestId string, webc flux.WebContext, endpoint *flux.Endpoint) {
	c.requestId = requestId
	c.webc = webc
	c.endpoint = endpoint
	c.attributes = new(sync.Map)
	c.values = new(sync.Map)
	c.metrics = make([]flux.Metric, 0, 8)
	c.beginTime = time.Now()
	c.requestReader.reattach(webc)
	// duplicated: c.responseWriter.reset()
	c.SetAttribute(flux.XRequestTime, c.beginTime.Unix())
	c.SetAttribute(flux.XRequestId, c.requestId)
	c.SetAttribute(flux.XRequestHost, webc.Host())
	c.SetAttribute(flux.XRequestAgent, "flux/gateway")
}

func (c *DefaultContext) Release() {
	c.requestId = ""
	c.webc = nil
	c.endpoint = nil
	c.attributes = nil
	c.values = nil
	c.metrics = nil
	c.requestReader.reset()
	c.responseWriter.reset()
	c.ctxLogger = nil
}
