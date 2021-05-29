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
	if err := checkjson(bytes); err != err {
		return emptyService, err
	}
	service := flux.Service{}
	if err := ext.JSONUnmarshal(bytes, &service); nil != err {
		return emptyService, fmt.Errorf("DECODE/MALFORMED/JSON: err: %w", err)
	}
	// 检查有效性
	if !service.IsValid() {
		return emptyService, errors.New("DECODE/MALFORMED/SERVICE")
	}
	return service, nil
}

func WrapServiceEvent(srv *flux.Service, etype remoting.EventType) (fxEvt flux.ServiceEvent, err error) {
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
