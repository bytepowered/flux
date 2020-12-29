package server

import (
	"github.com/bytepowered/flux"
	"strings"
)

func NewEchoEndpoints() []flux.HttpEndpointEvent {
	return []flux.HttpEndpointEvent{
		{
			EventType: flux.EventTypeAdded,
			Endpoint:  newEndpoint("get"),
		},
		{
			EventType: flux.EventTypeAdded,
			Endpoint:  newEndpoint("put"),
		},
		{
			EventType: flux.EventTypeAdded,
			Endpoint:  newEndpoint("post"),
		},
		{
			EventType: flux.EventTypeAdded,
			Endpoint:  newEndpoint("delete"),
		},
	}
}

func newEndpoint(method string) flux.Endpoint {
	return flux.Endpoint{
		Application: "fluxcore",
		Version:     "1.0",
		HttpPattern: "/debug/flux/echo/" + method,
		HttpMethod:  strings.ToUpper(method),
		Authorize:   false,
		Service: flux.BackendService{
			ServiceId: "flux.debug." + method,
			Interface: "flux.debug.EchoService",
			Method:    method,
			RpcProto:  flux.ProtoEcho,
		},
	}
}
