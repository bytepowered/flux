package registry

import (
	"github.com/bytepowered/flux"
)

type endpoint struct {
	Method  string
	Pattern string
}

// echoRegistry 实现Echo协议的注册中心
type echoRegistry int

func EchoRegistryFactory() flux.Registry {
	return new(echoRegistry)
}

// Id 返回Registry支持的协议作为ID标识
func (r *echoRegistry) Id() string {
	return "echo"
}

// 监听Metadata配置变化
func (r *echoRegistry) WatchEvents(outboundEvents chan<- flux.EndpointEvent) error {
	func(endpoints []endpoint) {
		for _, ep := range endpoints {
			outboundEvents <- r.toEndpointEvent(ep)
		}
	}([]endpoint{
		{Method: "GET", Pattern: "/debug/echo/get"},
		{Method: "POST", Pattern: "/debug/echo/post"},
		{Method: "DELETE", Pattern: "/debug/echo/delete"},
		{Method: "PUT", Pattern: "/debug/echo/put"},
	})
	return nil
}

func (r *echoRegistry) toEndpointEvent(ep endpoint) flux.EndpointEvent {
	return flux.EndpointEvent{
		Type:        flux.EndpointEventAdded,
		HttpMethod:  ep.Method,
		HttpPattern: ep.Pattern,
		Endpoint: flux.Endpoint{
			Version:        "v1",
			Protocol:       flux.ProtocolEcho,
			UpstreamUri:    ep.Pattern,
			UpstreamMethod: ep.Method,
			HttpPattern:    ep.Pattern,
			HttpMethod:     ep.Method,
			Arguments:      nil,
		},
	}
}
