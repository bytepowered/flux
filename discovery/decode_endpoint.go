package discovery

import (
	"errors"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/remoting"
)

var (
	emptyEndpoint      = flux.Endpoint{}
	emptyEndpointEvent = flux.EndpointEvent{}
)

func DecodeEndpointFunc(bytes []byte) (flux.Endpoint, error) {
	if err := checkjson(bytes); err != err {
		return emptyEndpoint, err
	}
	ep := flux.Endpoint{}
	if err := ext.JSONUnmarshal(bytes, &ep); nil != err {
		return emptyEndpoint, fmt.Errorf("DECODE/MALFORMED/JSON: err: %w", err)
	}
	// 检查Endpoint有效性
	if ep.HttpPattern == "" || ep.HttpMethod == "" {
		return emptyEndpoint, errors.New("DECODE/MALFORMED/ENDPOINT")
	}
	return ep, nil
}

func WrapEndpointEvent(ep *flux.Endpoint, etype remoting.EventType) (fxEvt flux.EndpointEvent, err error) {
	event := flux.EndpointEvent{Endpoint: *ep}
	switch etype {
	case remoting.EventTypeNodeAdd:
		event.EventType = flux.EventTypeAdded
	case remoting.EventTypeNodeDelete:
		event.EventType = flux.EventTypeRemoved
	case remoting.EventTypeNodeUpdate:
		event.EventType = flux.EventTypeUpdated
	default:
		return emptyEndpointEvent, fmt.Errorf("ENDPOINT:UNKNOWN_EVT_TYPE: type=%d", etype)
	}
	return event, nil
}
