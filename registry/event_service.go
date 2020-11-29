package registry

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/remoting"
)

var (
	_invalidBackendServiceEvent = flux.BackendServiceEvent{}
)

func toBackendServiceEvent(bytes []byte, etype remoting.EventType) (fxEvt flux.BackendServiceEvent, ok bool) {
	// Check json text
	size := len(bytes)
	if size < len("{\"k\":0}") || (bytes[0] != '[' && bytes[size-1] != '}') {
		logger.Infow("Invalid service event data.size", "data", string(bytes))
		return _invalidBackendServiceEvent, false
	}
	service := flux.BackendService{}
	json := ext.LoadSerializer(ext.TypeNameSerializerJson)
	if err := json.Unmarshal(bytes, &service); nil != err {
		logger.Warnw("Invalid service data",
			"event-type", etype, "data", string(bytes), "error", err)
		return _invalidBackendServiceEvent, false
	}
	logger.Infow("Received service event",
		"event-type", etype, "service-id", service.ServiceId, "data", string(bytes))
	// 检查有效性
	if !service.IsValid() {
		logger.Warnw("illegal backend service", "data", string(bytes))
		return _invalidBackendServiceEvent, false
	}
	event := flux.BackendServiceEvent{
		Service: service,
	}
	switch etype {
	case remoting.EventTypeNodeAdd:
		event.EventType = flux.EventTypeAdded
	case remoting.EventTypeNodeDelete:
		event.EventType = flux.EventTypeRemoved
	case remoting.EventTypeNodeUpdate:
		event.EventType = flux.EventTypeUpdated
	default:
		return _invalidBackendServiceEvent, false
	}
	return event, true
}
