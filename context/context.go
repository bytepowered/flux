package context

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"sync"
	"time"
)

var _ flux.Context = new(AttachableContext)

// Context接口实现
type AttachableContext struct {
	webc       flux.WebContext
	endpoint   *flux.Endpoint
	attributes *sync.Map
	variables  *sync.Map
	metrics    []flux.Metric
	startTime  time.Time
	request    *DefaultRequest
	response   *DefaultResponse
	ctxLogger  flux.Logger
}

func NewAttachableContext(webc flux.WebContext, endpoint *flux.Endpoint) *AttachableContext {
	a := NewAttachableContextWith(
		NewDefaultRequest(),
		NewDefaultResponse(),
		logger.SimpleLogger(),
	)
	a.attach(webc, endpoint)
	return a
}

func NewAttachableContextWith(request *DefaultRequest, response *DefaultResponse, logger flux.Logger) *AttachableContext {
	return &AttachableContext{
		request:   request,
		response:  response,
		ctxLogger: logger,
	}
}

func (c *AttachableContext) Request() flux.Request {
	return c.request
}

func (c *AttachableContext) Response() flux.Response {
	return c.response
}

func (c *AttachableContext) Endpoint() flux.Endpoint {
	return *(c.endpoint)
}

func (c *AttachableContext) BackendService() flux.BackendService {
	return c.endpoint.Service
}

func (c *AttachableContext) ServiceInterface() (proto, host, interfaceName, methodName string) {
	s := c.endpoint.Service
	return s.AttrRpcProto(), s.RemoteHost, s.Interface, s.Method
}

func (c *AttachableContext) ServiceProto() string {
	return c.endpoint.Service.AttrRpcProto()
}

func (c *AttachableContext) BackendServiceId() string {
	return c.endpoint.Service.ServiceID()
}

func (c *AttachableContext) Authorize() bool {
	return c.endpoint.AttrAuthorize()
}

func (c *AttachableContext) Method() string {
	return c.webc.Method()
}

func (c *AttachableContext) URI() string {
	return c.webc.URI()
}

func (c *AttachableContext) RequestId() string {
	return c.webc.RequestId()
}

func (c *AttachableContext) Attributes() map[string]interface{} {
	copied := make(map[string]interface{}, 16)
	c.attributes.Range(func(k, v interface{}) bool {
		copied[k.(string)] = v
		return true
	})
	return copied
}

func (c *AttachableContext) Attribute(key string, defval interface{}) interface{} {
	if v, ok := c.GetAttribute(key); ok {
		return v
	} else {
		return defval
	}
}

func (c *AttachableContext) SetAttribute(key string, value interface{}) {
	c.attributes.Store(key, value)
}

func (c *AttachableContext) GetAttribute(key string) (interface{}, bool) {
	v, ok := c.attributes.Load(key)
	return v, ok
}

func (c *AttachableContext) SetVariable(key string, value interface{}) {
	c.variables.Store(key, value)
}

func (c *AttachableContext) Variable(key string, defval interface{}) interface{} {
	if v, ok := c.GetVariable(key); ok {
		return v
	} else {
		return defval
	}
}

func (c *AttachableContext) GetVariable(key string) (interface{}, bool) {
	// first: Context Local Variables
	// then: WebContext Variables
	if lv, ok := c.variables.Load(key); ok {
		return lv, true
	} else if cv := c.webc.Variable(key); nil != cv {
		return cv, true
	} else {
		return nil, false
	}
}

func (c *AttachableContext) Context() context.Context {
	return c.webc.Context()
}

func (c *AttachableContext) StartAt() time.Time {
	return c.startTime
}

func (c *AttachableContext) Metrics() []flux.Metric {
	dist := make([]flux.Metric, len(c.metrics))
	copy(dist, c.metrics)
	return dist
}

func (c *AttachableContext) AddMetric(name string, elapsed time.Duration) {
	c.metrics = append(c.metrics, flux.Metric{
		Name: name, Elapsed: elapsed, Elapses: elapsed.String(),
	})
}

func (c *AttachableContext) SetLogger(logger flux.Logger) {
	c.ctxLogger = logger
}

func (c *AttachableContext) Logger() flux.Logger {
	return c.ctxLogger
}

func (c *AttachableContext) attach(webc flux.WebContext, endpoint *flux.Endpoint) *AttachableContext {
	c.webc = webc
	c.endpoint = endpoint
	c.attributes = new(sync.Map)
	c.variables = new(sync.Map)
	c.metrics = make([]flux.Metric, 0, 8)
	c.startTime = time.Now()
	c.request.attach(webc)
	c.SetAttribute(flux.XRequestTime, c.startTime.Unix())
	c.SetAttribute(flux.XRequestId, webc.RequestId())
	c.SetAttribute(flux.XRequestHost, webc.Host())
	c.SetAttribute(flux.XRequestAgent, "flux/gateway")
	return c
}
