package discovery

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/remoting"
)

var (
	invalidServiceEvent = flux.ServiceEvent{}
)

func NewServiceEvent(bytes []byte, etype remoting.EventType, node string) (fxEvt flux.ServiceEvent, ok bool) {
	// Check json text
	size := len(bytes)
	if size < len("{\"k\":0}") || (bytes[0] != '[' && bytes[size-1] != '}') {
		logger.Warnw("DISCOVERY:SERVICE:ILLEGAL_JSONSIZE", "data", string(bytes), "node", node)
		return invalidServiceEvent, false
	}
	service := flux.Service{}
	if err := ext.JSONUnmarshal(bytes, &service); nil != err {
		logger.Warnw("DISCOVERY:SERVICE:ILLEGAL_JSONFORMAT",
			"event-type", etype, "data", string(bytes), "error", err, "node", node)
		return invalidServiceEvent, false
	}
	// 检查有效性
	if !service.IsValid() {
		logger.Warnw("DISCOVERY:SERVICE:INVALID_VALUES", "service", service, "node", node)
		return invalidServiceEvent, false
	}
	// 兼容旧结构: before V0.10
	if len(service.Attributes) == 0 && service.Kind == "" {
		service.Attributes = []flux.Attribute{
			{Name: flux.ServiceAttrTagRpcProto, Value: flux.ProtoDubbo},
			{Name: flux.ServiceAttrTagRpcGroup, Value: ""},
			{Name: flux.ServiceAttrTagRpcVersion, Value: ""},
			{Name: flux.ServiceAttrTagRpcRetries, Value: "0"},
			{Name: flux.ServiceAttrTagRpcTimeout, Value: "10s"},
		}
	}

	event := flux.ServiceEvent{Service: service}
	switch etype {
	case remoting.EventTypeNodeAdd:
		event.EventType = flux.EventTypeAdded
	case remoting.EventTypeNodeDelete:
		event.EventType = flux.EventTypeRemoved
	case remoting.EventTypeNodeUpdate:
		event.EventType = flux.EventTypeUpdated
	default:
		return invalidServiceEvent, false
	}

	return event, true
}
