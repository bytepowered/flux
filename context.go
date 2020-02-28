package flux

const (
	XRequestId    = "X-Request-Id"
	XRequestTime  = "X-Request-Time"
	XRequestHost  = "X-Request-Host"
	XRequestAgent = "X-Request-Agent"
	XJwtSubject   = "X-Jwt-Subject"
	XJwtIssuer    = "X-Jwt-Issuer"
	XJwtToken     = "X-Jwt-Token"
)

// Context 定义每个请求的上下文环境
type Context interface {
	// RequestMethod 返回当前请求的Method
	RequestMethod() string

	// RequestUri 返回当前请求的URI
	RequestUri() string

	// RequestHost 返回当前请求的Host地址
	RequestHost() string

	// RequestId 返回当前请求的唯一ID
	RequestId() string

	// AttrValues 返回所有Attributes
	AttrValues() StringMap

	// AttrValue 获取指定name的Attribute，返回值和是否存在标识
	AttrValue(name string) (interface{}, bool)

	// SetAttrValue 向Context添加Attribute。
	SetAttrValue(name string, value interface{})

	// ScopedValue 获取当前请求范围的值
	ScopedValue(name string) (interface{}, bool)

	// SetScopedValue 设置当前请求范围的KV
	SetScopedValue(name string, value interface{})

	// Endpoint 返回请求路由定义的元数据
	Endpoint() Endpoint

	// ResponseWriter 返回响应数据接口
	ResponseWriter() ResponseWriter

	// RequestReader 返回请求数据接口
	RequestReader() RequestReader
}
