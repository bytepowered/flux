package server

import (
	"github.com/bytepowered/flux"
	"sync"
)

func NewBindEndpoint(endpoint *flux.Endpoint) *BindEndpoint {
	return &BindEndpoint{
		versions: map[string]*flux.Endpoint{
			endpoint.Version: endpoint,
		},
		RWMutex: new(sync.RWMutex),
	}
}

// Multi version Endpoint
type BindEndpoint struct {
	versions      map[string]*flux.Endpoint // 各版本数据
	*sync.RWMutex                           // 读写锁
}

// Find find endpoint by version
func (m *BindEndpoint) FindByVersion(version string) (*flux.Endpoint, bool) {
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

func (m *BindEndpoint) Update(version string, endpoint *flux.Endpoint) {
	m.Lock()
	m.versions[version] = endpoint
	m.Unlock()
}

func (m *BindEndpoint) Delete(version string) {
	m.Lock()
	delete(m.versions, version)
	m.Unlock()
}

func (m *BindEndpoint) RandomVersion() *flux.Endpoint {
	m.RLock()
	rv := m.random()
	m.RUnlock()
	return rv
}

func (m *BindEndpoint) random() *flux.Endpoint {
	for _, v := range m.versions {
		return v
	}
	return nil
}

func (m *BindEndpoint) ToSerializable() map[string]*flux.Endpoint {
	copies := make(map[string]*flux.Endpoint)
	m.RLock()
	for k, v := range m.versions {
		copies[k] = v
	}
	m.RUnlock()
	return copies
}
