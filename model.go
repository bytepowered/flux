package flux

import (
	"github.com/spf13/cast"
	"strings"
	"sync"
)

type (
	EventType int
)

// 路由元数据事件类型
const (
	EventTypeAdded = iota
	EventTypeUpdated
	EventTypeRemoved
)

const (
	// 从动态Path参数中获取
	ScopePath = "PATH"
	// 查询所有Path参数
	ScopePathMap = "PATH_MAP"
	// 从Query参数中获取
	ScopeQuery      = "QUERY"
	ScopeQueryMulti = "QUERY_MUL"
	// 获取全部Query参数
	ScopeQueryMap = "QUERY_MAP"
	// 只从Form表单参数参数列表中读取
	ScopeForm      = "FORM"
	ScopeFormMulti = "FORM_MUL"
	// 获取Form全部参数
	ScopeFormMap = "FORM_MAP"
	// 只从Query和Form表单参数参数列表中读取
	ScopeParam = "PARAM"
	// 只从Header参数中读取
	ScopeHeader = "HEADER"
	// 获取Header全部参数
	ScopeHeaderMap = "HEADER_MAP"
	// 获取Http Attributes的单个参数
	ScopeAttr = "ATTR"
	// 获取Http Attributes的Map结果
	ScopeAttrs = "ATTRS"
	// 获取Body数据
	ScopeBody = "BODY"
	// 获取Request元数据
	ScopeRequest = "REQUEST"
	// 自动查找数据源
	ScopeAuto = "AUTO"
)

const (
	// 原始参数类型：int,long...
	ArgumentTypePrimitive = "PRIMITIVE"
	// 复杂参数类型：POJO
	ArgumentTypeComplex = "COMPLEX"
	// JSONMap类型
	ArgumentTypeJSONMap = "JSONMAP"
)

// Support protocols
const (
	ProtoDubbo = "DUBBO"
	ProtoGRPC  = "GRPC"
	ProtoHttp  = "HTTP"
	ProtoEcho  = "ECHO"
)

// ServiceAttributes
const (
	ServiceAttrTagNotDefined = ""
	ServiceAttrTagRpcProto   = "rpcProto"
	ServiceAttrTagRpcGroup   = "rpcGroup"
	ServiceAttrTagRpcVersion = "rpcVersion"
	ServiceAttrTagRpcTimeout = "rpcTimeout"
	ServiceAttrTagRpcRetries = "rpcRetries"
)

// EndpointAttributes
const (
	EndpointAttrTagNotDefined = ""           // 默认的，未定义的属性
	EndpointAttrTagAuthorize  = "authorize"  // 标识Endpoint访问是否需要授权
	EndpointAttrTagListenerId = "listenerId" // 标识Endpoint绑定到哪个ListenServer服务
	EndpointAttrTagBizId      = "bizId"      // 标识Endpoint绑定到业务标识
)

// ArgumentAttributes
const (
	ArgumentAttributeTagDefault = "default" // 参数的默认值属性
)

type (
	// LookupFunc 参数值查找函数
	LookupFunc func(ctx *Context, scope, key string) (MTValue, error)
)

type (
	// Argument 定义Endpoint的参数结构元数据
	Argument struct {
		Name       string     `json:"name" yaml:"name"`             // 参数名称
		Type       string     `json:"type" yaml:"type"`             // 参数结构类型
		Class      string     `json:"class" yaml:"class"`           // 参数类型
		Generic    []string   `json:"generic" yaml:"generic"`       // 泛型类型
		HttpName   string     `json:"httpName" yaml:"httpName"`     // 映射Http的参数Key
		HttpScope  string     `json:"httpScope" yaml:"httpScope"`   // 映射Http参数值域
		Fields     []Argument `json:"fields" yaml:"fields"`         // 子结构字段
		Attributes Attributes `json:"attributes" yaml:"attributes"` // 属性列表
		// helper func
		ValueLoader   func() MTValue  `json:"-"`
		LookupFunc    LookupFunc      `json:"-"`
		ValueResolver MTValueResolver `json:"-"`
	}
)

// Attribute 定义服务的属性信息
type Attribute struct {
	Name  string      `json:"name" yaml:"name"`   // 属性名
	Value interface{} `json:"value" yaml:"value"` // 属性值
}

func (a Attribute) ToString() string {
	if values, ok := a.Value.([]interface{}); ok {
		if len(values) > 0 {
			return cast.ToString(values[0])
		} else {
			return ""
		}
	} else {
		return cast.ToString(a.Value)
	}
}

func (a Attribute) ToStringSlice() []string {
	return cast.ToStringSlice(a.Value)
}

func (a Attribute) ToInt() int {
	return cast.ToInt(a.Value)
}

func (a Attribute) ToBool() bool {
	return cast.ToBool(a.Value)
}

func (a Attribute) IsValid() bool {
	return a.Name != "" && a.Value != nil
}

// Attributes 定义属性列表
type Attributes []Attribute

// Single 查询单个属性
func (attrs Attributes) Single(name string) Attribute {
	v, _ := attrs.SingleEx(name)
	return v
}

func (attrs Attributes) SingleEx(name string) (Attribute, bool) {
	for _, attr := range attrs {
		if strings.EqualFold(attr.Name, name) {
			return attr, true
		}
	}
	return Attribute{}, false
}

func (attrs Attributes) Multiple(name string) []Attribute {
	out := make([]Attribute, 0, 2)
	for _, attr := range attrs {
		if strings.EqualFold(attr.Name, name) {
			out = append(out, attr)
		}
	}
	return out
}

func (attrs Attributes) Exists(name string) bool {
	for _, attr := range attrs {
		if strings.EqualFold(attr.Name, name) {
			return true
		}
	}
	return false
}

// Service 定义连接上游目标服务的信息
type Service struct {
	Kind        string     `json:"kind" yaml:"kind"`               // Service类型
	Application string     `json:"application" yaml:"application"` // 所属应用名
	AliasId     string     `json:"aliasId" yaml:"aliasId"`         // Service别名
	ServiceId   string     `json:"serviceId" yaml:"serviceId"`     // Service的标识ID
	Scheme      string     `json:"scheme" yaml:"scheme"`           // Service侧URL的Scheme
	Url         string     `json:"url" yaml:"url"`                 // Service侧的Host
	Interface   string     `json:"interface" yaml:"interface"`     // Service侧的URL/Interface
	Method      string     `json:"method" yaml:"method"`           // Service侧的方法
	Arguments   []Argument `json:"arguments" yaml:"arguments"`     // Service侧的参数结构
	Attributes  Attributes `json:"attributes" yaml:"attributes"`   // Service侧的属性列表
}

func (b Service) RpcProto() string {
	return b.Attributes.Single(ServiceAttrTagRpcProto).ToString()
}

func (b Service) RpcTimeout() string {
	return b.Attributes.Single(ServiceAttrTagRpcTimeout).ToString()
}

func (b Service) RpcGroup() string {
	return b.Attributes.Single(ServiceAttrTagRpcGroup).ToString()
}

func (b Service) RpcVersion() string {
	return b.Attributes.Single(ServiceAttrTagRpcVersion).ToString()
}

func (b Service) RpcRetries() string {
	return b.Attributes.Single(ServiceAttrTagRpcRetries).ToString()
}

// IsValid 判断服务配置是否有效；Interface+Method不能为空；
func (b Service) IsValid() bool {
	return b.Interface != "" && "" != b.Method
}

// HasArguments 判定是否有参数
func (b Service) HasArguments() bool {
	return len(b.Arguments) > 0
}

// ServiceID 构建标识当前服务的ID
func (b Service) ServiceID() string {
	return b.Interface + ":" + b.Method
}

// Endpoint 定义前端Http请求与后端RPC服务的端点元数据
type Endpoint struct {
	Kind        string     `json:"kind" yaml:"kind"`               // Endpoint类型
	Application string     `json:"application" yaml:"application"` // 所属应用名
	Version     string     `json:"version" yaml:"version"`         // 端点版本号
	HttpPattern string     `json:"httpPattern" yaml:"httpPattern"` // 映射Http侧的UriPattern
	HttpMethod  string     `json:"httpMethod" yaml:"httpMethod"`   // 映射Http侧的Method
	ServiceId   string     `json:"serviceId" yaml:"serviceId"`     // Service Id refer to Service
	Permissions []string   `json:"permissions" yaml:"permissions"` // 多组权限验证服务ID列表
	Attributes  Attributes `json:"attributes" yaml:"attributes"`   // 属性列表
	Service     Service    `json:"-"`
}

func (e *Endpoint) PermissionIds() []string {
	ids := make([]string, len(e.Permissions))
	copy(ids, e.Permissions)
	return ids
}

func (e *Endpoint) IsValid() bool {
	return e.HttpMethod != "" && e.HttpPattern != "" && e.ServiceId != ""
}

func (e *Endpoint) Authorize() bool {
	return e.Attributes.Single(EndpointAttrTagAuthorize).ToBool()
}

func (e *Endpoint) Attr(name string) Attribute {
	return e.attrs().Single(name)
}

func (e *Endpoint) AttrEx(name string) (Attribute, bool) {
	return e.attrs().SingleEx(name)
}

func (e *Endpoint) AttrExists(name string) bool {
	return e.attrs().Exists(name)
}

func (e *Endpoint) MultiAttrs(name string) Attributes {
	return e.attrs().Multiple(name)
}

func (e *Endpoint) attrs() Attributes {
	if e.Attributes == nil {
		return make(Attributes, 0)
	}
	return e.Attributes
}

// MVCEndpoint Multi version control Endpoint
type MVCEndpoint struct {
	versions      map[string]*Endpoint // 各版本数据
	*sync.RWMutex                      // 读写锁
}

func NewMVCEndpoint(endpoint *Endpoint) *MVCEndpoint {
	return &MVCEndpoint{
		versions: map[string]*Endpoint{
			endpoint.Version: endpoint,
		},
		RWMutex: new(sync.RWMutex),
	}
}

func (m *MVCEndpoint) IsEmpty() bool {
	m.RLock()
	defer m.RUnlock()
	return len(m.versions) == 0
}

// Lookup lookup by version, returns a copy endpoint,and a flag
func (m *MVCEndpoint) Lookup(version string) (Endpoint, bool) {
	m.RLock()
	defer m.RUnlock()
	size := len(m.versions)
	if 0 == size {
		return Endpoint{}, false
	}
	if "" == version || 1 == size {
		for _, ep := range m.versions {
			return m.dup(ep), true
		}
	}
	epv, ok := m.versions[version]
	if !ok {
		return Endpoint{}, false
	}
	return m.dup(epv), true
}

func (m *MVCEndpoint) dup(src *Endpoint) Endpoint {
	dup := *src
	return dup
}

func (m *MVCEndpoint) Update(version string, endpoint *Endpoint) {
	m.Lock()
	m.versions[version] = endpoint
	m.Unlock()
}

func (m *MVCEndpoint) Delete(version string) {
	m.Lock()
	delete(m.versions, version)
	m.Unlock()
}

func (m *MVCEndpoint) Random() Endpoint {
	m.RLock()
	defer m.RUnlock()
	for _, ep := range m.versions {
		return *ep
	}
	panic("SERVER:CRITICAL:ASSERT: <multi-endpoint> must not empty, on query random")
}

func (m *MVCEndpoint) Endpoints() []*Endpoint {
	m.RLock()
	copies := make([]*Endpoint, 0, len(m.versions))
	for _, ep := range m.versions {
		copies = append(copies, ep)
	}
	m.RUnlock()
	return copies
}

// EndpointEvent  定义从注册中心接收到的Endpoint数据变更
type EndpointEvent struct {
	EventType EventType
	Endpoint  Endpoint
}

// ServiceEvent  定义从注册中心接收到的Service定义数据变更
type ServiceEvent struct {
	EventType EventType
	Service   Service
}
