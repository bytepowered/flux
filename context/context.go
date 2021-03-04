package context

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/internal"
	"github.com/bytepowered/flux/logger"
	"sync"
	"time"
)

var _ flux.Context = new(AttachedContext)

var (
	pool = sync.Pool{
		New: func() interface{} {
			return &AttachedContext{
				request:   NewDefaultRequest(),
				response:  NewDefaultResponse(),
				ctxLogger: logger.SimpleLogger(),
			}
		},
	}
)

// Context接口实现
type AttachedContext struct {
	context    context.Context
	webc       flux.WebContext
	attributes map[string]interface{}
	variables  map[string]interface{}
	metrics    []flux.Metric
	startTime  time.Time
	request    *DefaultRequest
	response   *DefaultResponse
	ctxLogger  flux.Logger
}

func AcquireContext() *AttachedContext {
	return pool.Get().(*AttachedContext)
}

func ReleaseContext(c *AttachedContext) {
	pool.Put(c)
}

func New(webc flux.WebContext, endpoint *flux.Endpoint) *AttachedContext {
	ctx := AcquireContext()
	ctx.attach(webc, endpoint)
	return ctx
}

func (c *AttachedContext) Request() flux.Request {
	return c.request
}

func (c *AttachedContext) Response() flux.Response {
	return c.response
}

func (c *AttachedContext) Endpoint() flux.Endpoint {
	return *(c.context.Value(internal.ContextKeyRouteEndpoint).(*flux.Endpoint))
}

func (c *AttachedContext) BackendService() flux.BackendService {
	return c.endpoint().Service
}

func (c *AttachedContext) ServiceProto() string {
	return c.endpoint().Service.AttrRpcProto()
}

func (c *AttachedContext) BackendServiceId() string {
	return c.endpoint().Service.ServiceID()
}

func (c *AttachedContext) Authorize() bool {
	return c.endpoint().AttrAuthorize()
}

func (c *AttachedContext) Method() string {
	return c.webc.Method()
}

func (c *AttachedContext) URI() string {
	return c.webc.URI()
}

func (c *AttachedContext) RequestId() string {
	return c.webc.RequestId()
}

func (c *AttachedContext) Attributes() map[string]interface{} {
	copied := make(map[string]interface{}, len(c.attributes))
	for k, v := range c.attributes {
		copied[k] = v
	}
	return copied
}

func (c *AttachedContext) Attribute(key string, defval interface{}) interface{} {
	if v, ok := c.attributes[key]; ok {
		return v
	} else {
		return defval
	}
}

func (c *AttachedContext) SetAttribute(key string, value interface{}) {
	if nil == c.attributes {
		c.attributes = make(map[string]interface{}, 16)
	}
	c.attributes[key] = value
}

func (c *AttachedContext) GetAttribute(key string) (interface{}, bool) {
	v, ok := c.attributes[key]
	return v, ok
}

func (c *AttachedContext) SetVariable(key string, value interface{}) {
	if nil == c.variables {
		c.variables = make(map[string]interface{}, 16)
	}
	c.variables[key] = value
}

func (c *AttachedContext) Variable(key string, defval interface{}) interface{} {
	if v, ok := c.lookupVar(key); ok {
		return v
	} else {
		return defval
	}
}

func (c *AttachedContext) GetVariable(key string) (interface{}, bool) {
	return c.lookupVar(key)
}

func (c *AttachedContext) lookupVar(key string) (interface{}, bool) {
	// first: Context Local Variables
	// then: WebContext Variables
	if lv, ok := c.variables[key]; ok {
		return lv, true
	} else if cv := c.webc.Variable(key); nil != cv {
		return cv, true
	} else {
		return nil, false
	}
}

func (c *AttachedContext) Context() context.Context {
	return c.context
}

func (c *AttachedContext) StartAt() time.Time {
	return c.startTime
}

func (c *AttachedContext) Metrics() []flux.Metric {
	dist := make([]flux.Metric, len(c.metrics))
	copy(dist, c.metrics)
	return dist
}

func (c *AttachedContext) AddMetric(name string, elapsed time.Duration) {
	c.metrics = append(c.metrics, flux.Metric{
		Name: name, Elapsed: elapsed, Elapses: elapsed.String(),
	})
}

func (c *AttachedContext) SetLogger(logger flux.Logger) {
	c.ctxLogger = logger
}

func (c *AttachedContext) Logger() flux.Logger {
	return c.ctxLogger
}

func (c *AttachedContext) endpoint() *flux.Endpoint {
	return c.context.Value(internal.ContextKeyRouteEndpoint).(*flux.Endpoint)
}

func (c *AttachedContext) attach(webc flux.WebContext, endpoint *flux.Endpoint) *AttachedContext {
	c.context = context.WithValue(webc.Context(), internal.ContextKeyRouteEndpoint, endpoint)
	c.webc = webc
	c.attributes = nil
	c.variables = nil
	c.metrics = nil
	c.ctxLogger = logger.SimpleLogger()
	c.startTime = time.Now()
	c.request.reset(webc)
	c.response.reset()
	c.SetAttribute(flux.XRequestTime, c.startTime.Unix())
	c.SetAttribute(flux.XRequestId, webc.RequestId())
	c.SetAttribute(flux.XRequestHost, webc.Host())
	c.SetAttribute(flux.XRequestAgent, "flux/gateway")
	return c
}
