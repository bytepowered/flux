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
	RequestMethod() string                        //返回当前请求的Method
	RequestUri() string                           //返回当前请求的URI
	RequestPath() string                          //返回当前请求的URI的路径
	RequestHost() string                          //返回当前请求的Host地址
	RequestId() string                            // 返回当前请求的唯一ID
	ResponseWriter() ResponseWriter               // 返回响应数据接口
	RequestReader() RequestReader                 // 返回请求数据接口
	Endpoint() Endpoint                           // 返回请求路由定义的元数据
	EndpointProtoName() string                    // 返回Endpoint的协议名称
	Attributes() map[string]interface{}           // 返回所有Attributes键值对
	GetAttribute(name string) (interface{}, bool) // 获取指定name的Attribute，返回值和是否存在标识
	SetAttribute(name string, value interface{})  // 向Context添加Attribute键值对。
	GetValue(name string) (interface{}, bool)     // 获取当前请求范围的值
	SetValue(name string, value interface{})      // 设置当前请求范围的KV
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
	switch strings.ToUpper(parts[0]) {
	case ScopeQuery:
		return req.QueryValue(parts[1])
	case ScopeForm:
		return req.FormValue(parts[1])
	case ScopePath:
		return req.PathValue(parts[1])
	case ScopeHeader:
		return req.HeaderValue(parts[1])
	case ScopeAttr:
		v, _ := ctx.GetAttribute(parts[1])
		return v
	default:
		return nil
	}
}
