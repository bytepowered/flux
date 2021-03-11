package ext

import (
	flux2 "github.com/bytepowered/flux/flux-node"
	"sync"
)

var (
	endpoints = new(sync.Map)
)

func RegisterEndpoint(key string, endpoint *flux2.Endpoint) *flux2.MultiEndpoint {
	mve := flux2.NewMultiEndpoint(endpoint)
	endpoints.Store(key, mve)
	return mve
}

func EndpointByKey(key string) (*flux2.MultiEndpoint, bool) {
	ep, ok := endpoints.Load(key)
	if ok {
		return ep.(*flux2.MultiEndpoint), true
	}
	return nil, false
}

func Endpoints() map[string]*flux2.MultiEndpoint {
	out := make(map[string]*flux2.MultiEndpoint, 32)
	endpoints.Range(func(key, value interface{}) bool {
		out[key.(string)] = value.(*flux2.MultiEndpoint)
		return true
	})
	return out
}
