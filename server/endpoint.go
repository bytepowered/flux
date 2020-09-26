package server

import (
	"github.com/bytepowered/flux"
	"sync"
)

func NewMultiVersionEndpoint(endpoint *flux.Endpoint) *MultiVersionEndpoint {
	return &MultiVersionEndpoint{
		versioned: map[string]*flux.Endpoint{
			endpoint.Version: endpoint,
		},
		mutex: new(sync.RWMutex),
	}
}

// Multi version Endpoint
type MultiVersionEndpoint struct {
	versioned map[string]*flux.Endpoint // 各版本数据
	mutex     *sync.RWMutex             // 读写锁
}

// Find find endpoint by version
func (m *MultiVersionEndpoint) FindByVersion(version string) (*flux.Endpoint, bool) {
	m.mutex.RLock()
	if "" == version || 1 == len(m.versioned) {
		rv := m.random()
		m.mutex.RUnlock()
		return rv, nil != rv
	}
	v, ok := m.versioned[version]
	m.mutex.RUnlock()
	return v, ok
}

func (m *MultiVersionEndpoint) Update(version string, endpoint *flux.Endpoint) {
	m.mutex.Lock()
	m.versioned[version] = endpoint
	m.mutex.Unlock()
}

func (m *MultiVersionEndpoint) Delete(version string) {
	m.mutex.Lock()
	delete(m.versioned, version)
	m.mutex.Unlock()
}

func (m *MultiVersionEndpoint) RandomVersion() *flux.Endpoint {
	m.mutex.RLock()
	rv := m.random()
	m.mutex.RUnlock()
	return rv
}

func (m *MultiVersionEndpoint) random() *flux.Endpoint {
	for _, v := range m.versioned {
		return v
	}
	return nil
}

func (m *MultiVersionEndpoint) ToSerializable() map[interface{}]interface{} {
	copies := make(map[interface{}]interface{})
	m.mutex.RLock()
	for k, v := range m.versioned {
		copies[k] = v
	}
	m.mutex.RUnlock()
	return copies
}
