package registry

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/remoting"
)

var (
	invalidHttpEndpointEvent = flux.HttpEndpointEvent{}
)

type CompatibleEndpoint struct {
	flux.Endpoint
	Authorize bool `json:"authorize"` // 此端点是否需要授权
}

func NewEndpointEvent(bytes []byte, etype remoting.EventType) (fxEvt flux.HttpEndpointEvent, ok bool) {
	// Check json text
	size := len(bytes)
	if size < len("{\"k\":0}") || (bytes[0] != '[' && bytes[size-1] != '}') {
		logger.Infow("Invalid endpoint event data.size", "data", string(bytes))
		return invalidHttpEndpointEvent, false
	}
	comp := CompatibleEndpoint{}
	if err := ext.JSONUnmarshal(bytes, &comp); nil != err {
		logger.Warnw("invalid endpoint data",
			"event-type", etype, "data", string(bytes), "error", err)
		return invalidHttpEndpointEvent, false
	}
	logger.Infow("Received endpoint event",
		"event-type", etype, "method", comp.HttpMethod, "pattern", comp.HttpPattern, "data", string(bytes))
	// 兼容旧协议数据格式
	fixesServiceAttributes(&comp.Service)
	fixesServiceAttributes(&comp.Permission)
	if len(comp.Attributes) == 0 {
		comp.Attributes = []flux.Attribute{
			{
				Tag:   flux.EndpointAttrTagAuthorize,
				Name:  "Authorize",
				Value: comp.Authorize,
			},
		}
	}
	// 检查有效性
	if !comp.IsValid() {
		logger.Warnw("illegal http-metadata", "data", string(bytes))
		return invalidHttpEndpointEvent, false
	}
	if !comp.Service.IsValid() {
		logger.Warnw("illegal service", "service", comp.Service, "data", string(bytes))
		return invalidHttpEndpointEvent, false
	}

	event := flux.HttpEndpointEvent{
		Endpoint: comp.Endpoint,
	}
	switch etype {
	case remoting.EventTypeNodeAdd:
		event.EventType = flux.EventTypeAdded
	case remoting.EventTypeNodeDelete:
		event.EventType = flux.EventTypeRemoved
	case remoting.EventTypeNodeUpdate:
		event.EventType = flux.EventTypeUpdated
	default:
		return invalidHttpEndpointEvent, false
	}
	return event, true
}
