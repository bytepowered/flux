package flux

const (
	KeyConfigRootEndpointRegistry   = "EndpointDiscovery"
	KeyConfigEndpointDiscoveryProto = "discovery-proto"
)

// EndpointDiscovery Endpoint注册元数据事件监听
// 监听接收元数据中心的配置变化
type EndpointDiscovery interface {
	OnEndpointChanged() (<-chan HttpEndpointEvent, error)
	OnServiceChanged() (<-chan BackendServiceEvent, error)
}
