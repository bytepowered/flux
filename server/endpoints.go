package server

import (
	"github.com/bytepowered/flux"
	"sync"
)

var (
	endpoints = new(sync.Map)
)

func SelectMultiEndpoint(key string) (*MultiEndpoint, bool) {
	ep, ok := endpoints.Load(key)
	if ok {
		return ep.(*MultiEndpoint), true
	}
	return nil, false
}

func RegisterMultiEndpoint(key string, endpoint *flux.Endpoint) *MultiEndpoint {
	mve := newMultiEndpoint(endpoint)
	endpoints.Store(key, mve)
	return mve
}

func LoadEndpoints() map[string]*MultiEndpoint {
	out := make(map[string]*MultiEndpoint, 32)
	endpoints.Range(func(key, value interface{}) bool {
		out[key.(string)] = value.(*MultiEndpoint)
		return true
	})
	return out
}

// Multi version Endpoint
type MultiEndpoint struct {
	endpoint      map[string]*flux.Endpoint // 各版本数据
	*sync.RWMutex                           // 读写锁
}

func newMultiEndpoint(endpoint *flux.Endpoint) *MultiEndpoint {
	return &MultiEndpoint{
		endpoint: map[string]*flux.Endpoint{
			endpoint.Version: endpoint,
		},
		RWMutex: new(sync.RWMutex),
	}
}

// Find find endpoint by version
func (m *MultiEndpoint) FindByVersion(version string) (*flux.Endpoint, bool) {
	m.RLock()
	if "" == version || 1 == len(m.endpoint) {
		rv := m.random()
		m.RUnlock()
		return rv, nil != rv
	}
	v, ok := m.endpoint[version]
	m.RUnlock()
	return v, ok
}

func (m *MultiEndpoint) Update(version string, endpoint *flux.Endpoint) {
	m.Lock()
	m.endpoint[version] = endpoint
	m.Unlock()
}

func (m *MultiEndpoint) Delete(version string) {
	m.Lock()
	delete(m.endpoint, version)
	m.Unlock()
}

func (m *MultiEndpoint) RandomVersion() *flux.Endpoint {
	m.RLock()
	rv := m.random()
	m.RUnlock()
	return rv
}

func (m *MultiEndpoint) random() *flux.Endpoint {
	for _, v := range m.endpoint {
		return v
	}
	return nil
}

func (m *MultiEndpoint) ToSerializable() map[string]*flux.Endpoint {
	copies := make(map[string]*flux.Endpoint)
	m.RLock()
	for k, v := range m.endpoint {
		copies[k] = v
	}
	m.RUnlock()
	return copies
}
