package flux

import (
	"go.uber.org/zap"
	"time"
)

const (
	XRequestId    = "X-Request-Id"
	XRequestTime  = "X-Request-Time"
	XRequestHost  = "X-Request-Host"
	XRequestAgent = "X-Request-Agent"
)

type (
	// Context 定义每个请求的上下文环境。包含当前请求的Endpoint、属性、请求Context等。
	// Context是一个与请求生命周期相同的临时上下文容器，它在请求被接受处理时创建，在请求被处理完成时销毁。
	// Attributes是Context的属性参数，它可以在多个组件之间传递，并且可以传递到后端Dubbo/Http/gRPC等服务。
	Context struct {
		ServerWebContext
		endpoint   *Endpoint
		attributes map[string]interface{}
		metrics    []TraceMetric
		startTime  time.Time
		ctxLogger  Logger
	}

	// OnContextHookFunc 用于WebContext与Context的交互勾子；
	// 在每个请求被路由执行时，在创建Context后被调用。
	OnContextHookFunc func(ServerWebContext, *Context)

	// OnBeforeFilterHookFunc 在Filter执行前被调用的勾子函数
	OnBeforeFilterHookFunc func(*Context, []Filter)

	// OnBeforeTransportHookFunc 在Transporter执行前被调用的勾子函数
	OnBeforeTransportHookFunc func(*Context, Transporter)

	// LookupScopedValueFunc 参数值查找函数
	LookupScopedValueFunc func(ctx *Context, scope, key string) (MTValue, error)
)

// NewContext 构建新的Context实例。
func NewContext() *Context {
	return &Context{
		attributes: make(map[string]interface{}, 16),
		metrics:    make([]TraceMetric, 0, 16),
	}
}

// Reset 重置Context，重新与 ServerWebContext 关联，绑定新的 Endpoint
// Note：此函数由内部框架调用，一般不作为业务使用。
func (c *Context) Reset(webex ServerWebContext, endpoint *Endpoint) {
	c.ServerWebContext = webex
	c.endpoint = endpoint
	c.ctxLogger = zap.S()
	c.startTime = time.Now()
	c.metrics = c.metrics[:0]
	for k := range c.attributes {
		delete(c.attributes, k)
	}
}

// Application 返回当前Endpoint对应的应用名
func (c *Context) Application() string {
	return c.endpoint.Application
}

// Endpoint 返回当前请求路由定义的Endpoint元数据
func (c *Context) Endpoint() *Endpoint {
	return c.endpoint
}

// Service 返回Service信息
func (c *Context) Service() Service {
	return c.endpoint.Service
}

// ServiceID 返回Endpoint Service的服务标识
func (c *Context) ServiceID() string {
	return c.endpoint.Service.ServiceID()
}

// Attribute 获取指定key的Attribute。如果不存在，返回默认值；
func (c *Context) Attribute(key string, defval interface{}) interface{} {
	if v, ok := c.GetAttribute(key); ok {
		return v
	} else {
		return defval
	}
}

// Attributes 返回所有Attributes键值对；只读；
// 搜索位置及顺序：
// 1. Context自身的Attributes；【HIGH】
// 2. Endpoint的Attributes；【LOW】
func (c *Context) Attributes() map[string]interface{} {
	out := make(map[string]interface{}, len(c.attributes))
	for _, attr := range c.endpoint.Attributes {
		out[attr.Name] = attr.Value
	}
	for k, v := range c.attributes {
		out[k] = v
	}
	return out
}

// GetAttribute 获取指定key的Attribute，返回值和是否存在标识
// 搜索位置及顺序：
// 1. Context自身的Attributes；【HIGH】
// 2. Endpoint的Attributes；【LOW】
func (c *Context) GetAttribute(key string) (interface{}, bool) {
	v, ok := c.attributes[key]
	if !ok {
		if attr, aok := c.endpoint.AttributeEx(key); aok {
			return attr.Value, true
		}
	}
	return v, ok
}

// SetAttribute 向Context添加Attribute键值对
func (c *Context) SetAttribute(key string, value interface{}) {
	c.attributes[key] = value
}

// StartAt 返回Http请求起始的服务器时间
func (c *Context) StartAt() time.Time {
	return c.startTime
}

// AddMetric 添加路由耗时统计节点
func (c *Context) AddMetric(name string, elapsed time.Duration) {
	c.metrics = append(c.metrics, TraceMetric{
		Name: name, Elapses: elapsed.String(),
	})
}

// Metrics 返回请求路由的的统计数据
func (c *Context) Metrics() []TraceMetric {
	dist := make([]TraceMetric, len(c.metrics))
	copy(dist, c.metrics)
	return dist
}

// SetLogger 添加Context范围的Logger。
// 通常是将关联一些追踪字段的Logger设置为ContextLogger
func (c *Context) SetLogger(logger Logger) {
	c.ctxLogger = logger
}

// Logger 返回Context范围的Logger。
func (c *Context) Logger() Logger {
	return c.ctxLogger
}

// TraceMetric 请求路由的的统计数据
type TraceMetric struct {
	Name    string `json:"name"`
	Elapses string `json:"elapses"`
}
