package flux

// Service内置注解
const (
	ServiceAnnotationRpcGroup   = "flux.go/rpc.group"
	ServiceAnnotationRpcVersion = "flux.go/rpc.version"
	ServiceAnnotationRpcTimeout = "flux.go/rpc.timeout"
	ServiceAnnotationRpcRetries = "flux.go/rpc.retries"
)

const (
	// ServiceArgumentTypePrimitive 原始参数类型：int,long...
	ServiceArgumentTypePrimitive = "PRIMITIVE"
	// ServiceArgumentTypeComplex 复杂参数类型：POJO
	ServiceArgumentTypeComplex = "COMPLEX"
)

const (
	ServiceArgumentAnnotationDefault = "default" // 参数的默认值属性
)

// ServiceSpec 定义连接上游目标服务的信息
type ServiceSpec struct {
	Kind        string                `json:"kind" yaml:"kind"`               // Service类型
	AliasId     string                `json:"aliasId" yaml:"aliasId"`         // Service别名
	Url         string                `json:"url" yaml:"url"`                 // Service侧的URL
	Protocol    string                `json:"protocol" yaml:"protocol"`       // Service侧后端协议
	Interface   string                `json:"interface" yaml:"interface"`     // Service侧的Interface
	Method      string                `json:"method" yaml:"method"`           // Service侧的Method
	Arguments   []ServiceArgumentSpec `json:"arguments" yaml:"arguments"`     // Service侧的参数结构
	Annotations Annotations           `json:"annotations" yaml:"annotations"` // Service侧的注解列表
}

// Annotation 获取指定名称的注解，如果注解不存在，返回空注解。
func (s ServiceSpec) Annotation(name string) NamedValueSpec {
	return s.Annotations.Get(name)
}

// IsValid 判断服务配置是否有效；
// 1. Interface, Method 不能为空；
// 2. 包含Proto协议；
func (s ServiceSpec) IsValid() bool {
	return s.Interface != "" && s.Method != "" && s.Protocol != ""
}

// ServiceID 构建标识当前服务的ID
func (s ServiceSpec) ServiceID() string {
	if s.Interface == "" || s.Method == "" {
		return ""
	}
	return s.Interface + ":" + s.Method
}

// ServiceArgumentSpec 定义Endpoint的参数结构元数据
type ServiceArgumentSpec struct {
	Name         string                `json:"name" yaml:"name"`               // 参数名称
	StructType   string                `json:"type" yaml:"type"`               // 参数结构类型
	ClassType    string                `json:"class" yaml:"class"`             // 参数类型
	GenericTypes []string              `json:"generic" yaml:"generic"`         // 泛型类型
	HttpName     string                `json:"httpName" yaml:"httpName"`       // 映射Http的参数Key
	HttpScope    string                `json:"httpScope" yaml:"httpScope"`     // 映射Http参数值域
	Fields       []ServiceArgumentSpec `json:"fields" yaml:"fields"`           // 子结构字段
	Annotations  Annotations           `json:"annotations" yaml:"annotations"` // 注解列表
	extends      map[interface{}]interface{}
}

func (ptr *ServiceArgumentSpec) SetExtends(key, val interface{}) {
	if ptr.extends == nil {
		ptr.extends = make(map[interface{}]interface{}, 4)
	}
	ptr.extends[key] = val
}

func (ptr *ServiceArgumentSpec) GetExtends(key interface{}) (interface{}, bool) {
	if ptr.extends == nil {
		return nil, false
	}
	v, ok := ptr.extends[key]
	return v, ok
}
