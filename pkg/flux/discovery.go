package flux

import (
	"context"
)

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
)

type EventType int

// 路由元数据事件类型
const (
	EventTypeAdded = iota
	EventTypeUpdated
	EventTypeRemoved
)

var eventTypeNames = map[int]string{
	-1:               "EventType:Unknown",
	EventTypeAdded:   "EventType:Added",
	EventTypeUpdated: "EventType:Updated",
	EventTypeRemoved: "EventType:Removed",
}

func EventTypeName(evtType int) string {
	if n, ok := eventTypeNames[evtType]; ok {
		return n
	}
	return eventTypeNames[-1]
}

// EndpointEvent  定义从注册中心接收到的Endpoint数据变更
type EndpointEvent struct {
	EventType EventType
	Endpoint  EndpointSpec
}

// ServiceEvent  定义从注册中心接收到的Service定义数据变更
type ServiceEvent struct {
	EventType EventType
	Service   ServiceSpec
}
