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

func toEndpointEvent(bytes []byte, etype remoting.EventType) (fxEvt flux.HttpEndpointEvent, ok bool) {
	// Check json text
	size := len(bytes)
	if size < len("{\"k\":0}") || (bytes[0] != '[' && bytes[size-1] != '}') {
		logger.Infow("Invalid endpoint event data.size", "data", string(bytes))
		return invalidHttpEndpointEvent, false
	}
	endpoint := flux.Endpoint{}
	if err := ext.JSONUnmarshal(bytes, &endpoint); nil != err {
		logger.Warnw("invalid endpoint data",
			"event-type", etype, "data", string(bytes), "error", err)
		return invalidHttpEndpointEvent, false
	}
	logger.Infow("Received endpoint event",
		"event-type", etype, "method", endpoint.HttpMethod, "pattern", endpoint.HttpPattern, "data", string(bytes))
	// 检查有效性
	if !endpoint.IsValid() {
		logger.Warnw("illegal http-pattern", "data", string(bytes))
		return invalidHttpEndpointEvent, false
	}
	if !endpoint.Service.IsValid() {
		logger.Warnw("illegal service", "service", endpoint.Service, "data", string(bytes))
		return invalidHttpEndpointEvent, false
	}
	event := flux.HttpEndpointEvent{
		Endpoint: endpoint,
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
