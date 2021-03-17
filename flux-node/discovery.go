package flux

// EndpointDiscovery Endpoint注册元数据事件监听
// 监听接收元数据中心的配置变化
type EndpointDiscovery interface {
	// Id 返回标识当前服务标识
	Id() string

	// WatchEndpoints 监听HttpEndpoint注册事件
	WatchEndpoints(events chan<- EndpointEvent) error

	// WatchServices 监听TransporterService注册事件
	WatchServices(events chan<- ServiceEvent) error
}
