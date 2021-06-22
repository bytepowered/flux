package discovery

import (
	"errors"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/remoting"
)

var (
	emptyEndpoint      = flux.EndpointSpec{}
	emptyEndpointEvent = flux.EndpointEvent{}
)

func DecodeEndpointFunc(bytes []byte) (flux.EndpointSpec, error) {
	if err := VerifyJSON(bytes); err != nil {
		return emptyEndpoint, err
	}
	ep := flux.EndpointSpec{}
	if err := ext.JSONUnmarshal(bytes, &ep); nil != err {
		return emptyEndpoint, fmt.Errorf("DECODE:UNMARSHAL:JSON/err: %w", err)
	}
	if ep.Annotations == nil {
		ep.Annotations = make(flux.Annotations, 0)
	}
	if ep.Attributes == nil {
		ep.Attributes = make(flux.Attributes, 0)
	}
	// 检查有效性
	if !ep.IsValid() {
		return emptyEndpoint, errors.New("DECODE:VERIFY:ENDPOINT/invalid")
	}
	return ep, nil
}

func ToEndpointEvent(ep *flux.EndpointSpec, etype remoting.NodeEventType) (fxEvt flux.EndpointEvent, err error) {
	event := flux.EndpointEvent{Endpoint: *ep}
	switch etype {
	case remoting.EventTypeNodeAdd:
		event.EventType = flux.EventTypeAdded
	case remoting.EventTypeNodeDelete:
		event.EventType = flux.EventTypeRemoved
	case remoting.EventTypeNodeUpdate:
		event.EventType = flux.EventTypeUpdated
	default:
		return emptyEndpointEvent, fmt.Errorf("ENDPOINT:UNKNOWN_EVT_TYPE: type=%s", etype)
	}
	return event, nil
}
