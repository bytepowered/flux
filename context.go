package flux

import "strings"

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

	// RequestPath 返回当前请求的URI的路径
	RequestPath() string

	// RequestHost 返回当前请求的Host地址
	RequestHost() string

	// RequestId 返回当前请求的唯一ID
	RequestId() string

	// ResponseWriter 返回响应数据接口
	ResponseWriter() ResponseWriter

	// RequestReader 返回请求数据接口
	RequestReader() RequestReader

	// Endpoint 返回请求路由定义的元数据
	Endpoint() Endpoint

	// Attributes 返回所有Attributes键值对
	Attributes() map[string]interface{}

	// GetAttribute 获取指定name的Attribute，返回值和是否存在标识
	GetAttribute(name string) (interface{}, bool)

	// SetAttribute 向Context添加Attribute键值对。
	SetAttribute(name string, value interface{})

	// GetValue 获取当前请求范围的值
	GetValue(name string) (interface{}, bool)

	// SetValue 设置当前请求范围的KV
	SetValue(name string, value interface{})
}

// LookupValue 搜索Lookup指定域的值。支持：
// 1. query:<name>
// 2. form:<name>
// 3. path:<name>
// 4. header:<name>
// 5. attr:<name>
func LookupValue(lookup string, ctx Context) interface{} {
	if "" == lookup || nil == ctx {
		return nil
	}
	req := ctx.RequestReader()
	parts := strings.Split(lookup, ":")
	if len(parts) == 1 {
		return req.HeaderValue(parts[0])
	}
	switch strings.ToLower(parts[0]) {
	case "query":
		return req.QueryValue(parts[1])
	case "form":
		return req.FormValue(parts[1])
	case "path":
		return req.PathValue(parts[1])
	case "header":
		return req.HeaderValue(parts[1])
	case "attr":
		v, _ := ctx.GetAttribute(parts[1])
		return v
	default:
		return nil
	}
}
