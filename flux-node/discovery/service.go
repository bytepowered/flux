package discovery

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/bytepowered/flux/flux-node/logger"
	"github.com/bytepowered/flux/flux-node/remoting"
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
	EnsureServiceAttrs(&service)

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

// EnsureServiceAttrs 兼容旧协议数据格式
func EnsureServiceAttrs(service *flux.Service) *flux.Service {
	if len(service.Attributes) == 0 {
		service.Attributes = []flux.Attribute{
			{Name: flux.ServiceAttrTagRpcProto, Value: service.AttrRpcProto},
			{Name: flux.ServiceAttrTagRpcGroup, Value: service.AttrRpcGroup},
			{Name: flux.ServiceAttrTagRpcVersion, Value: service.AttrRpcVersion},
			{Name: flux.ServiceAttrTagRpcRetries, Value: service.AttrRpcRetries},
			{Name: flux.ServiceAttrTagRpcTimeout, Value: service.AttrRpcTimeout},
		}
	}
	return service
}
