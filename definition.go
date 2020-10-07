package flux

import "fmt"

type (
	EventType int
)

// 路由元数据事件类型
const (
	EndpointEventAdded = iota
	EndpointEventUpdated
	EndpointEventRemoved
)

const (
	// 自动查找数据源
	ScopeAuto = "AUTO"
	// 获取Http Attributes的单个参数
	ScopeAttr = "ATTR"
	// 获取Http Attributes的Map结果
	ScopeAttrs = "ATTRS"
	// 只从Form表单参数参数列表中读取
	ScopeForm = "FORM"
	// 只从Header参数中读取
	ScopeHeader = "HEADER"
	// 只从Query和Form表单参数参数列表中读取
	ScopeParam = "PARAM"
	// 从动态Path参数中获取
	ScopePath = "PATH"
	// 从Query参数中获取
	ScopeQuery = "QUERY"
	// 获取Body数据
	ScopeBody = "BODY"
)

const (
	// 原始参数类型：int,long...
	ArgumentTypePrimitive = "PRIMITIVE"
	// 复杂参数类型：POJO
	ArgumentTypeComplex = "COMPLEX"
)

// Support protocols
const (
	ProtoDubbo = "DUBBO"
	ProtoGRPC  = "GRPC"
	ProtoHttp  = "HTTP"
)

// Argument 定义Endpoint的参数结构元数据
type Argument struct {
	Name        string     `json:"argName"`     // 参数名称
	Type        string     `json:"argType"`     // 参数结构类型
	TypeClass   string     `json:"typeClass"`   // 参数类型
	TypeGeneric []string   `json:"typeGeneric"` // 泛型类型
	HttpName    string     `json:"httpName"`    // 映射Http的参数Key
	HttpScope   string     `json:"httpScope"`   // 映射Http参数值域
	HttpValue   Valuer     `json:"-"`           // 参数值
	Fields      []Argument `json:"fields"`      // 子结构字段
}

// Endpoint 定义前端Http请求与后端RPC服务的端点元数据
type Endpoint struct {
	Application    string                 `json:"application"`       // 所属应用名
	Version        string                 `json:"version"`           // 端点版本号
	RpcGroup       string                 `json:"rpcGroup"`          // rpc接口分组
	RpcVersion     string                 `json:"rpcVersion"`        // rpc接口版本
	RpcTimeout     string                 `json:"rpcTimeout"`        // RPC调用超时
	RpcRetries     string                 `json:"rpcRetries,string"` // RPC调用重试
	Authorize      bool                   `json:"authorize"`         // 此端点是否需要授权
	UpstreamProto  string                 `json:"protocol"`          // 定义Upstream侧的协议
	UpstreamHost   string                 `json:"upstreamHost"`      // 定义Upstream侧的Host
	UpstreamUri    string                 `json:"upstreamUri"`       // 定义Upstream侧的URL
	UpstreamMethod string                 `json:"upstreamMethod"`    // 定义Upstream侧的方法
	HttpPattern    string                 `json:"httpPattern"`       // 映射Http侧的UriPattern
	HttpMethod     string                 `json:"httpMethod"`        // 映射Http侧的Method
	Arguments      []Argument             `json:"arguments"`         // 参数结构
	Permission     Permission             `json:"permission"`        // 权限验证定义
	Extensions     map[string]interface{} `json:"extensions"`        // 扩展信息
}

// Permission 后端RPC服务的权限验证的元数据
type Permission struct {
	UpstreamProto  string     `json:"protocol"`       // 定义Upstream侧的协议
	UpstreamHost   string     `json:"upstreamHost"`   // 定义Upstream侧的Host
	UpstreamUri    string     `json:"upstreamUri"`    // 定义Upstream侧的URL
	UpstreamMethod string     `json:"upstreamMethod"` // 定义Upstream侧的方法
	Arguments      []Argument `json:"arguments"`      // 参数结构
}

func (p Permission) IsValid() bool {
	return "" != p.UpstreamProto && "" != p.UpstreamUri && "" != p.UpstreamMethod && len(p.Arguments) > 0
}

// NewServiceKey 构建标识一个Service的Key字符串
func NewServiceKey(proto, host, method, uri string) string {
	return fmt.Sprintf("%s@%s:%s/%s", proto, host, method, uri)
}

// EndpointEvent  定义从注册中心接收到的Endpoint数据变更
type EndpointEvent struct {
	EventType   EventType
	HttpMethod  string `json:"method"`
	HttpPattern string `json:"pattern"`
	Endpoint    Endpoint
}

/// Value

func NewWrapValue(v interface{}) Valuer {
	return &ValueWrapper{
		value: v,
	}
}

type ValueWrapper struct {
	value interface{}
}

func (v *ValueWrapper) Value() interface{} {
	return v.value
}

func (v *ValueWrapper) SetValue(value interface{}) {
	v.value = value
}
