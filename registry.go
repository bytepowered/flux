package flux

const (
	KeyConfigRootEndpointRegistry  = "EndpointRegistry"
	KeyConfigEndpointRegistryProto = "registry-proto"
)

// EndpointRegistry Endpoint注册元数据事件监听
// 监听接收元数据中心的配置变化
type EndpointRegistry interface {
	WatchHttpEndpoints() (<-chan HttpEndpointEvent, error)
	WatchBackendServices() (<-chan BackendServiceEvent, error)
}
