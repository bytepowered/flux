package server

import (
	"github.com/bytepowered/flux"
	"sync"
)

var (
	_endpoints = new(sync.Map)
)

func SelectMVEndpoint(key string) (*MVEndpoint, bool) {
	ep, ok := _endpoints.Load(key)
	if ok {
		return ep.(*MVEndpoint), true
	}
	return nil, false
}

func RegisterMVEndpoint(key string, endpoint *flux.Endpoint) *MVEndpoint {
	mve := newMVEndpoint(endpoint)
	_endpoints.Store(key, mve)
	return mve
}

func LoadEndpoints() map[string]*MVEndpoint {
	out := make(map[string]*MVEndpoint, 32)
	_endpoints.Range(func(key, value interface{}) bool {
		out[key.(string)] = value.(*MVEndpoint)
		return true
	})
	return out
}

// Multi version Endpoint
type MVEndpoint struct {
	versions      map[string]*flux.Endpoint // 各版本数据
	*sync.RWMutex                           // 读写锁
}

func newMVEndpoint(endpoint *flux.Endpoint) *MVEndpoint {
	return &MVEndpoint{
		versions: map[string]*flux.Endpoint{
			endpoint.Version: endpoint,
		},
		RWMutex: new(sync.RWMutex),
	}
}

// Find find endpoint by version
func (m *MVEndpoint) FindByVersion(version string) (*flux.Endpoint, bool) {
	m.RLock()
	if "" == version || 1 == len(m.versions) {
		rv := m.random()
		m.RUnlock()
		return rv, nil != rv
	}
	v, ok := m.versions[version]
	m.RUnlock()
	return v, ok
}

func (m *MVEndpoint) Update(version string, endpoint *flux.Endpoint) {
	m.Lock()
	m.versions[version] = endpoint
	m.Unlock()
}

func (m *MVEndpoint) Delete(version string) {
	m.Lock()
	delete(m.versions, version)
	m.Unlock()
}

func (m *MVEndpoint) RandomVersion() *flux.Endpoint {
	m.RLock()
	rv := m.random()
	m.RUnlock()
	return rv
}

func (m *MVEndpoint) random() *flux.Endpoint {
	for _, v := range m.versions {
		return v
	}
	return nil
}

func (m *MVEndpoint) ToSerializable() map[string]*flux.Endpoint {
	copies := make(map[string]*flux.Endpoint)
	m.RLock()
	for k, v := range m.versions {
		copies[k] = v
	}
	m.RUnlock()
	return copies
}
