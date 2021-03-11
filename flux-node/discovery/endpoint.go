package discovery

import (
	flux2 "github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/bytepowered/flux/flux-node/logger"
	"github.com/bytepowered/flux/flux-node/remoting"
)

var (
	invalidHttpEndpointEvent = flux2.HttpEndpointEvent{}
)

type CompatibleEndpoint struct {
	flux2.Endpoint
	// Deprecated
	Authorize bool `json:"authorize"` // 此端点是否需要授权
}

func NewEndpointEvent(bytes []byte, etype remoting.EventType, node string) (fxEvt flux2.HttpEndpointEvent, ok bool) {
	// Check json text
	size := len(bytes)
	if size < len("{\"k\":0}") || (bytes[0] != '[' && bytes[size-1] != '}') {
		logger.Warnw("DISCOVERY:ENDPOINT:ILLEGAL_JSONSIZE", "data", string(bytes), "node", node)
		return invalidHttpEndpointEvent, false
	}
	comp := CompatibleEndpoint{}
	if err := ext.JSONUnmarshal(bytes, &comp); nil != err {
		logger.Warnw("DISCOVERY:ENDPOINT:ILLEGAL_JSONFORMAT",
			"event-type", etype, "data", string(bytes), "error", err, "node", node)
		return invalidHttpEndpointEvent, false
	}
	// 检查有效性
	if !comp.IsValid() {
		logger.Warnw("DISCOVERY:ENDPOINT:INVALID_VALUES", "data", string(bytes), "node", node)
		return invalidHttpEndpointEvent, false
	}
	// 兼容旧结构
	if len(comp.Attributes) == 0 {
		comp.Attributes = []flux2.Attribute{
			{Name: flux2.EndpointAttrTagAuthorize, Value: comp.Authorize},
		}
	}
	EnsureServiceAttrs(&comp.Service)
	EnsureServiceAttrs(&comp.Permission)

	event := flux2.HttpEndpointEvent{Endpoint: comp.Endpoint}
	switch etype {
	case remoting.EventTypeNodeAdd:
		event.EventType = flux2.EventTypeAdded
	case remoting.EventTypeNodeDelete:
		event.EventType = flux2.EventTypeRemoved
	case remoting.EventTypeNodeUpdate:
		event.EventType = flux2.EventTypeUpdated
	default:
		return invalidHttpEndpointEvent, false
	}
	return event, true
}
