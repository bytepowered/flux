package context

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"sync"
	"time"
)

var _ flux.Context = new(DefaultContext)

// Context接口实现
type DefaultContext struct {
	requestId  string
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

func DefaultContextFactory() flux.Context {
	return &DefaultContext{
		response:  NewDefaultResponse(),
		request:   NewDefaultRequest(),
		ctxLogger: logger.SimpleLogger(),
	}
}

func (c *DefaultContext) Request() flux.Request {
	return c.request
}

func (c *DefaultContext) Response() flux.Response {
	return c.response
}

func (c *DefaultContext) Endpoint() flux.Endpoint {
	return *(c.endpoint)
}

func (c *DefaultContext) BackendService() flux.BackendService {
	return c.endpoint.Service
}

func (c *DefaultContext) ServiceInterface() (proto, host, interfaceName, methodName string) {
	s := c.endpoint.Service
	return s.AttrRpcProto(), s.RemoteHost, s.Interface, s.Method
}

func (c *DefaultContext) ServiceProto() string {
	return c.endpoint.Service.AttrRpcProto()
}

func (c *DefaultContext) BackendServiceId() string {
	return c.endpoint.Service.ServiceID()
}

func (c *DefaultContext) Authorize() bool {
	return c.endpoint.AttrAuthorize()
}

func (c *DefaultContext) Method() string {
	return c.webc.Method()
}

func (c *DefaultContext) URI() string {
	return c.webc.URI()
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

func (c *DefaultContext) Attribute(key string, defval interface{}) interface{} {
	if v, ok := c.GetAttribute(key); ok {
		return v
	} else {
		return defval
	}
}

func (c *DefaultContext) SetAttribute(key string, value interface{}) {
	c.attributes.Store(key, value)
}

func (c *DefaultContext) GetAttribute(key string) (interface{}, bool) {
	v, ok := c.attributes.Load(key)
	return v, ok
}

func (c *DefaultContext) SetVariable(key string, value interface{}) {
	c.variables.Store(key, value)
}

func (c *DefaultContext) Variable(key string, defval interface{}) interface{} {
	if v, ok := c.GetVariable(key); ok {
		return v
	} else {
		return defval
	}
}

func (c *DefaultContext) GetVariable(key string) (interface{}, bool) {
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

func (c *DefaultContext) Context() context.Context {
	return c.webc.Context()
}

func (c *DefaultContext) Metrics() []flux.Metric {
	dist := make([]flux.Metric, len(c.metrics))
	copy(dist, c.metrics)
	return dist
}

func (c *DefaultContext) SetLogger(logger flux.Logger) {
	c.ctxLogger = logger
}

func (c *DefaultContext) Logger() flux.Logger {
	return c.ctxLogger
}

func (c *DefaultContext) StartAt() time.Time {
	return c.startTime
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
	c.variables = new(sync.Map)
	c.metrics = make([]flux.Metric, 0, 8)
	c.startTime = time.Now()
	c.request.reattach(webc)
	// duplicated: c.response.reset()
	c.SetAttribute(flux.XRequestTime, c.startTime.Unix())
	c.SetAttribute(flux.XRequestId, c.requestId)
	c.SetAttribute(flux.XRequestHost, webc.Host())
	c.SetAttribute(flux.XRequestAgent, "flux/gateway")
}

func (c *DefaultContext) Release() {
	c.requestId = ""
	c.webc = nil
	c.endpoint = nil
	c.attributes = nil
	c.variables = nil
	c.metrics = nil
	c.request.reset()
	c.response.reset()
	c.ctxLogger = nil
}
