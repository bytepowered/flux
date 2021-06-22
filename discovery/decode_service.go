package discovery

import (
	"errors"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/remoting"
)

var (
	emptyService      = flux.Service{}
	emptyServiceEvent = flux.ServiceEvent{}
)

func DecodeServiceFunc(bytes []byte) (flux.Service, error) {
	if err := VerifyJSON(bytes); err != nil {
		return emptyService, err
	}
	service := flux.Service{}
	if err := ext.JSONUnmarshal(bytes, &service); nil != err {
		return emptyService, fmt.Errorf("DECODE:UNMARSHAL:JSON/err: %w", err)
	}
	// ensure
	if service.Annotations == nil {
		service.Annotations = make(flux.Annotations, 0)
	}
	// 检查有效性
	if !service.Valid() {
		return emptyService, errors.New("DECODE:VERIFY:SERVICE/invalid")
	}
	return service, nil
}

func ToServiceEvent(srv *flux.Service, etype remoting.NodeEventType) (fxEvt flux.ServiceEvent, err error) {
	event := flux.ServiceEvent{Service: *srv}
	switch etype {
	case remoting.EventTypeNodeAdd:
		event.EventType = flux.EventTypeAdded
	case remoting.EventTypeNodeDelete:
		event.EventType = flux.EventTypeRemoved
	case remoting.EventTypeNodeUpdate:
		event.EventType = flux.EventTypeUpdated
	default:
		return emptyServiceEvent, fmt.Errorf("DISCOVERY:SERVICE:UNKNOWN_EVT_TYPE: type=%d", etype)
	}

	return event, nil
}
