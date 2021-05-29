package flux

import "context"

// EndpointDiscovery Endpoint注册元数据事件监听
// 监听接收元数据中心的配置变化
type (
	EndpointDiscovery interface {
		// Id 返回标识当前服务标识
		Id() string

		// WatchEndpoints 监听HttpEndpoint注册事件
		WatchEndpoints(ctx context.Context, events chan<- EndpointEvent) error

		// WatchServices 监听TransporterService注册事件
		WatchServices(ctx context.Context, events chan<- ServiceEvent) error
	}

	// DiscoveryDecodeServiceFunc 将原始数据解码为Service事件
	DiscoveryDecodeServiceFunc func(bytes []byte) (service Service, err error)

	// DiscoveryDecodeEndpointFunc 将原始数据解码为Service事件
	DiscoveryDecodeEndpointFunc func(bytes []byte) (service Endpoint, err error)
)
