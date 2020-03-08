package internal

import (
	"github.com/bytepowered/flux"
	"sync"
)

func NewMultiVersionEndpoint() *MultiVersionEndpoint {
	return &MultiVersionEndpoint{
		data: new(sync.Map),
	}
}

// Multi version Endpoint
type MultiVersionEndpoint struct {
	data *sync.Map
}

func (m MultiVersionEndpoint) Get(version string) (*flux.Endpoint, bool) {
	v, ok := m.data.Load(version)
	if ok {
		return v.(*flux.Endpoint), true
	} else {
		return nil, false
	}
}

func (m MultiVersionEndpoint) Update(version string, endpoint *flux.Endpoint) {
	m.data.Store(version, endpoint)
}

func (m MultiVersionEndpoint) Delete(version string) {
	m.data.Delete(version)
}

func (m MultiVersionEndpoint) ToSerializableMap() map[interface{}]interface{} {
	amap := make(map[interface{}]interface{})
	m.data.Range(func(key, value interface{}) bool {
		amap[key] = value
		return false
	})
	return amap
}
