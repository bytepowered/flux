package context

import (
	"context"
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/internal"
	"github.com/bytepowered/flux/flux-node/logger"
	"sync"
	"time"
)

var _ flux.Context = new(AttacheContext)

var (
	pool = sync.Pool{
		New: func() interface{} {
			return &AttacheContext{
				request:   NewWebRequest(),
				response:  NewWebResponse(),
				ctxLogger: logger.SimpleLogger(),
			}
		},
	}
)

// Context接口实现
type AttacheContext struct {
	context    context.Context
	exchange   flux.WebExchange
	attributes map[string]interface{}
	variables  map[string]interface{}
	metrics    []flux.Metric
	startTime  time.Time
	request    *WebRequest
	response   *WebResponse
	ctxLogger  flux.Logger
}

func AcquireContext() *AttacheContext {
	return pool.Get().(*AttacheContext)
}

func ReleaseContext(c *AttacheContext) {
	pool.Put(c)
}

func New(webex flux.WebExchange, endpoint *flux.Endpoint) *AttacheContext {
	ctx := AcquireContext()
	ctx.attach(webex, endpoint)
	return ctx
}

func (c *AttacheContext) Request() flux.Request {
	return c.request
}

func (c *AttacheContext) Response() flux.Response {
	return c.response
}

func (c *AttacheContext) Endpoint() flux.Endpoint {
	return *c.endpoint()
}

func (c *AttacheContext) Application() string {
	return c.endpoint().Application
}

func (c *AttacheContext) BackendService() flux.BackendService {
	return c.endpoint().Service
}

func (c *AttacheContext) ServiceProto() string {
	return c.endpoint().Service.AttrRpcProto()
}

func (c *AttacheContext) BackendServiceId() string {
	return c.endpoint().Service.ServiceID()
}

func (c *AttacheContext) Authorize() bool {
	return c.endpoint().AttrAuthorize()
}

func (c *AttacheContext) Method() string {
	return c.exchange.Method()
}

func (c *AttacheContext) URI() string {
	return c.exchange.URI()
}

func (c *AttacheContext) RequestId() string {
	return c.exchange.RequestId()
}

func (c *AttacheContext) Attributes() map[string]interface{} {
	copied := make(map[string]interface{}, len(c.attributes))
	for k, v := range c.attributes {
		copied[k] = v
	}
	return copied
}

func (c *AttacheContext) Attribute(key string, defval interface{}) interface{} {
	if v, ok := c.attributes[key]; ok {
		return v
	} else {
		return defval
	}
}

func (c *AttacheContext) SetAttribute(key string, value interface{}) {
	if nil == c.attributes {
		c.attributes = make(map[string]interface{}, 16)
	}
	c.attributes[key] = value
}

func (c *AttacheContext) GetAttribute(key string) (interface{}, bool) {
	v, ok := c.attributes[key]
	return v, ok
}

func (c *AttacheContext) SetVariable(key string, value interface{}) {
	if nil == c.variables {
		c.variables = make(map[string]interface{}, 16)
	}
	c.variables[key] = value
}

func (c *AttacheContext) Variable(key string, defval interface{}) interface{} {
	if v, ok := c.lookupVar(key); ok {
		return v
	} else {
		return defval
	}
}

func (c *AttacheContext) GetVariable(key string) (interface{}, bool) {
	return c.lookupVar(key)
}

func (c *AttacheContext) lookupVar(key string) (interface{}, bool) {
	// first: Context Local Variables
	// then: WebExchange Variables
	if lv, ok := c.variables[key]; ok {
		return lv, true
	} else if cv := c.exchange.Variable(key); nil != cv {
		return cv, true
	} else {
		return nil, false
	}
}

func (c *AttacheContext) Context() context.Context {
	return c.context
}

func (c *AttacheContext) StartAt() time.Time {
	return c.startTime
}

func (c *AttacheContext) Metrics() []flux.Metric {
	dist := make([]flux.Metric, len(c.metrics))
	copy(dist, c.metrics)
	return dist
}

func (c *AttacheContext) AddMetric(name string, elapsed time.Duration) {
	c.metrics = append(c.metrics, flux.Metric{
		Name: name, Elapsed: elapsed, Elapses: elapsed.String(),
	})
}

func (c *AttacheContext) SetLogger(logger flux.Logger) {
	c.ctxLogger = logger
}

func (c *AttacheContext) Logger() flux.Logger {
	return c.ctxLogger
}

func (c *AttacheContext) endpoint() *flux.Endpoint {
	return c.context.Value(internal.ContextKeyRouteEndpoint).(*flux.Endpoint)
}

func (c *AttacheContext) attach(webex flux.WebExchange, endpoint *flux.Endpoint) *AttacheContext {
	c.context = context.WithValue(webex.Context(), internal.ContextKeyRouteEndpoint, endpoint)
	c.exchange = webex
	c.attributes = nil
	c.variables = nil
	c.metrics = nil
	c.ctxLogger = logger.SimpleLogger()
	c.startTime = time.Now()
	c.request.reset(webex)
	c.response.reset()
	return c
}
