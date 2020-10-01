package flux

const (
	KeyConfigRootEndpointRegistry = "Registry"
	KeyConfigEndpointRegistryId   = "registry-id"
)

// EndpointRegistry Endpoint注册元数据事件监听
type EndpointRegistry interface {
	// 监听接收元数据中心的配置变化
	Watch() (<-chan EndpointEvent, error)
}
