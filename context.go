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

// Context 定义每个请求的上下文环境
type Context struct {
	ServerWebContext
	endpoint   *Endpoint
	attributes map[string]interface{}
	metrics    []Metric
	startTime  time.Time
	ctxLogger  Logger
}

func NewContext() *Context {
	return &Context{
		attributes: make(map[string]interface{}, 16),
		metrics:    make([]Metric, 0, 16),
	}
}

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
	if v, ok := c.GetVariable(key); ok {
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
		if attr, aok := c.endpoint.GetAttrEx(key); aok {
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
	c.metrics = append(c.metrics, Metric{
		Name: name, Elapsed: elapsed, Elapses: elapsed.String(),
	})
}

// LoadMetrics 返回请求路由的的统计数据
func (c *Context) Metrics() []Metric {
	dist := make([]Metric, len(c.metrics))
	copy(dist, c.metrics)
	return dist
}

// GetLogger 添加Context范围的Logger。
// 通常是将关联一些追踪字段的Logger设置为ContextLogger
func (c *Context) SetLogger(logger Logger) {
	c.ctxLogger = logger
}

// GetLogger 返回Context范围的Logger。
func (c *Context) Logger() Logger {
	return c.ctxLogger
}

// Metrics 请求路由的的统计数据
type Metric struct {
	Name    string        `json:"name"`
	Elapsed time.Duration `json:"elapsed"`
	Elapses string        `json:"elapses"`
}
