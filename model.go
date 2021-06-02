package flux

import (
	"fmt"
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
	// ScopePath 从动态Path参数中获取
	ScopePath = "PATH"
	// ScopePathMap 查询所有Path参数
	ScopePathMap = "PATH_MAP"
	// ScopeQuery 从Query参数中获取
	ScopeQuery      = "QUERY"
	ScopeQueryMulti = "QUERY_MUL"
	// ScopeQueryMap 获取全部Query参数
	ScopeQueryMap = "QUERY_MAP"
	// ScopeForm 只从Form表单参数参数列表中读取
	ScopeForm      = "FORM"
	ScopeFormMulti = "FORM_MUL"
	// ScopeFormMap 获取Form全部参数
	ScopeFormMap = "FORM_MAP"
	// ScopeParam 只从Query和Form表单参数参数列表中读取
	ScopeParam = "PARAM"
	// ScopeHeader 只从Header参数中读取
	ScopeHeader = "HEADER"
	// ScopeHeaderMap 获取Header全部参数
	ScopeHeaderMap = "HEADER_MAP"
	// ScopeAttr 获取Http Attributes的单个参数
	ScopeAttr = "ATTR"
	// ScopeAttrs 获取Http Attributes的Map结果
	ScopeAttrs = "ATTRS"
	// ScopeBody 获取Body数据
	ScopeBody = "BODY"
	// ScopeRequest 获取Request元数据
	ScopeRequest = "REQUEST"
	// ScopeAuto 自动查找数据源
	ScopeAuto = "AUTO"
)

const (
	// ArgumentTypePrimitive 原始参数类型：int,long...
	ArgumentTypePrimitive = "PRIMITIVE"
	// ArgumentTypeComplex 复杂参数类型：POJO
	ArgumentTypeComplex = "COMPLEX"
	// ArgumentTypeJSONMap JSONMap类型
	ArgumentTypeJSONMap = "JSONMAP"
)

// Support protocols
const (
	ProtoDubbo = "DUBBO"
	ProtoGRPC  = "GRPC"
	ProtoHttp  = "HTTP"
	ProtoEcho  = "ECHO"
	ProtoInApp = "INAPP"
)

// ServiceAttributes
const (
	ServiceAttrTagNotDefined = ""
	ServiceAttrTagRpcProto   = "rpc:proto"
	ServiceAttrTagRpcGroup   = "rpc:group"
	ServiceAttrTagRpcVersion = "rpc:version"
	ServiceAttrTagRpcTimeout = "rpc:timeout"
	ServiceAttrTagRpcRetries = "rpc:retries"
)

// EndpointAttributes
const (
	EndpointAttrTagNotDefined = ""               // 默认的，未定义的属性
	EndpointAttrTagAuthorize  = "tag:authorize"  // 标识Endpoint访问是否需要授权
	EndpointAttrTagListenerId = "tag:listenerId" // 标识Endpoint绑定到哪个ListenServer服务
	EndpointAttrTagBizId      = "tag:bizId"      // 标识Endpoint绑定到业务标识
	EndpointAttrTagPermission = "tag:permission"
)

const (
	// ArgumentAttributeTagDefault ArgumentAttributes
	ArgumentAttributeTagDefault = "default" // 参数的默认值属性
)

type (
	// MTValueLoaderFunc 参值直接加载函数
	MTValueLoaderFunc func() MTValue
	// MTValueLookupFunc 参数值查找函数
	MTValueLookupFunc func(ctx *Context, scope, key string) (MTValue, error)
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
		ValueLoader   MTValueLoaderFunc `json:"-"`
		LookupFunc    MTValueLookupFunc `json:"-"`
		ValueResolver MTValueResolver   `json:"-"`
	}
)

// Resolve 解析Argument参数值
func (a Argument) Resolve(ctx *Context) (interface{}, error) {
	return resolve(ctx, &a)
}

// Resolvep 解析Argument参数值
func (a *Argument) Resolvep(ctx *Context) (interface{}, error) {
	return resolve(ctx, a)
}

// Resolve 解析Argument参数值
func resolve(ctx *Context, a *Argument) (interface{}, error) {
	if nil == a.ValueResolver {
		return nil, fmt.Errorf("ValueResolver is nil, name: %s", a.Name)
	}
	// 1: Value loader
	if nil != a.ValueLoader {
		mtv := a.ValueLoader()
		return a.ValueResolver(mtv, a.Class, a.Generic)
	}
	// 2: Lookup
	if nil == a.LookupFunc {
		return nil, fmt.Errorf("MTValueLookupFunc is nil, name: %s", a.Name)
	}
	// 3: Single value
	if len(a.Fields) == 0 {
		mtv, err := a.LookupFunc(ctx, a.HttpScope, a.HttpName)
		if nil != err {
			return nil, err
		}
		if !mtv.Valid && a.Attributes != nil {
			if attr, ok := a.Attributes.SingleEx(ArgumentAttributeTagDefault); ok {
				mtv = NewStringMTValue(attr.ToString())
			}
		}
		return a.ValueResolver(mtv, a.Class, a.Generic)
	}
	// 4: POJO Values
	sm := make(map[string]interface{}, len(a.Fields))
	sm["class"] = a.Class
	for _, field := range a.Fields {
		if fv, err := field.Resolve(ctx); nil != err {
			return nil, err
		} else {
			sm[field.Name] = fv
		}
	}
	return sm, nil
}

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

// SingleEx 查询单个属性，并返回是否存在标识
func (attrs Attributes) SingleEx(name string) (Attribute, bool) {
	for _, attr := range attrs {
		if strings.EqualFold(attr.Name, name) {
			return attr, true
		}
	}
	return Attribute{}, false
}

// Multiple 查询多个同名属性
func (attrs Attributes) Multiple(name string) Attributes {
	out := make(Attributes, 0, 2)
	for _, attr := range attrs {
		if strings.EqualFold(attr.Name, name) {
			out = append(out, attr)
		}
	}
	return out
}

// Exists 判定属性名是否存在
func (attrs Attributes) Exists(name string) bool {
	for _, attr := range attrs {
		if strings.EqualFold(attr.Name, name) {
			return true
		}
	}
	return false
}

// Values 返回属性列表的值
func (attrs Attributes) Values() []interface{} {
	out := make([]interface{}, len(attrs))
	for i, a := range attrs {
		out[i] = a.Value
	}
	return out
}

// Strings 返回属性列表的值
func (attrs Attributes) Strings() []string {
	out := make([]string, len(attrs))
	for i, a := range attrs {
		out[i] = a.ToString()
	}
	return out
}

// Service 定义连接上游目标服务的信息
type Service struct {
	Kind       string     `json:"kind" yaml:"kind"`             // Service类型
	AliasId    string     `json:"aliasId" yaml:"aliasId"`       // Service别名
	Scheme     string     `json:"scheme" yaml:"scheme"`         // Service侧URL的Scheme
	Url        string     `json:"url" yaml:"url"`               // Service侧的Host
	Interface  string     `json:"interface" yaml:"interface"`   // Service侧的URL/Interface
	Method     string     `json:"method" yaml:"method"`         // Service侧的方法
	Arguments  []Argument `json:"arguments" yaml:"arguments"`   // Service侧的参数结构
	Attributes Attributes `json:"attributes" yaml:"attributes"` // Service侧的属性列表
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
	Attributes  Attributes `json:"attributes" yaml:"attributes"`   // 属性列表
	ServiceId   string     `json:"serviceId" yaml:"serviceId"`     // 上游/后端服务ServiceId
	Service     Service    `json:"-"`                              // 上游/后端服务
}

func (e *Endpoint) IsValid() bool {
	return e.HttpMethod != "" && e.HttpPattern != "" &&
		e.AttrExists(ServiceAttrTagRpcProto)
}

// Deprecated Use Attribute instead
func (e *Endpoint) Attr(name string) Attribute {
	return e.Attribute(name)
}

// Attribute 获取指定名称的属性，如果属性不存在，返回空属性。
func (e *Endpoint) Attribute(name string) Attribute {
	return e.ensureAttributes().Single(name)
}

// Deprecated Use AttributeEx instead
func (e *Endpoint) AttrEx(name string) (Attribute, bool) {
	return e.AttributeEx(name)
}

// Attribute 获取指定名称的属性，如果属性不存在，返回空属性，以及是否存在标识位
func (e *Endpoint) AttributeEx(name string) (Attribute, bool) {
	return e.ensureAttributes().SingleEx(name)
}

// Deprecated Use AttributeExists instead
func (e *Endpoint) AttrExists(name string) bool {
	return e.AttributeExists(name)
}

// AttributeExists 判断指定名称的属性是否存在
func (e *Endpoint) AttributeExists(name string) bool {
	return e.ensureAttributes().Exists(name)
}

// Deprecated Use MultiAttributes instead
func (e *Endpoint) MultiAttrs(name string) Attributes {
	return e.MultiAttributes(name)
}

// MultiAttributes 获取指定属性名的多个属性列表
func (e *Endpoint) MultiAttributes(name string) Attributes {
	return e.ensureAttributes().Multiple(name)
}

func (e *Endpoint) ensureAttributes() Attributes {
	if e.Attributes == nil {
		e.Attributes = make(Attributes, 0)
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

// IsEmpty 判断多版本控制Endpoint是否为空
func (m *MVCEndpoint) IsEmpty() bool {
	m.RLock()
	defer m.RUnlock()
	return len(m.versions) == 0
}

// Lookup 按指定版本事情查找Endpoint。返回Endpoint的复制数据。
// 如果有且仅有一个版本，则直接返回此Endpoint，不比较版本号是否匹配。
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

// Update 更新指定版本号的Endpoint元数据
func (m *MVCEndpoint) Update(version string, endpoint *Endpoint) {
	m.Lock()
	m.versions[version] = endpoint
	m.Unlock()
}

// Delete 删除指定版本号的元数据
func (m *MVCEndpoint) Delete(version string) {
	m.Lock()
	delete(m.versions, version)
	m.Unlock()
}

// Random 随机读取一个版本的元数据。
// 注意：必须保证随机读取版本时存在非空元数据，否则会报错panic。
func (m *MVCEndpoint) Random() Endpoint {
	m.RLock()
	defer m.RUnlock()
	for _, ep := range m.versions {
		return *ep
	}
	panic("SERVER:CRITICAL:ASSERT: <multi-endpoint> must not empty, call by random query func")
}

// Endpoint 获取当前多版本控制器的全部Endpoint元数据列表
func (m *MVCEndpoint) Endpoints() []*Endpoint {
	m.RLock()
	copies := make([]*Endpoint, 0, len(m.versions))
	for _, ep := range m.versions {
		copies = append(copies, ep)
	}
	m.RUnlock()
	return copies
}

func (m *MVCEndpoint) dup(src *Endpoint) Endpoint {
	dup := *src
	return dup
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
