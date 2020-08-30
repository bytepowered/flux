package internal

import (
	"github.com/bytepowered/flux"
	"sync"
)

func NewMultiVersionEndpoint(endpoint *flux.Endpoint) *MultiVersionEndpoint {
	return &MultiVersionEndpoint{
		versionMap: map[string]*flux.Endpoint{
			endpoint.Version: endpoint,
		},
		rwmu: new(sync.RWMutex),
	}
}

// Multi version Endpoint
type MultiVersionEndpoint struct {
	versionMap map[string]*flux.Endpoint // 各版本数据
	rwmu       *sync.RWMutex             // 读写锁
}

// Find find endpoint by version
func (m *MultiVersionEndpoint) FindByVersion(version string) (*flux.Endpoint, bool) {
	m.rwmu.RLock()
	if "" == version || 1 == len(m.versionMap) {
		rv := m.random()
		m.rwmu.RUnlock()
		return rv, nil != rv
	}
	v, ok := m.versionMap[version]
	m.rwmu.RUnlock()
	return v, ok
}

func (m *MultiVersionEndpoint) Update(version string, endpoint *flux.Endpoint) {
	m.rwmu.Lock()
	m.versionMap[version] = endpoint
	m.rwmu.Unlock()
}

func (m *MultiVersionEndpoint) Delete(version string) {
	m.rwmu.Lock()
	delete(m.versionMap, version)
	m.rwmu.Unlock()
}

func (m *MultiVersionEndpoint) RandomVersion() *flux.Endpoint {
	m.rwmu.RLock()
	rv := m.random()
	m.rwmu.RUnlock()
	return rv
}

func (m *MultiVersionEndpoint) random() *flux.Endpoint {
	for _, v := range m.versionMap {
		return v
	}
	return nil
}

func (m *MultiVersionEndpoint) ToSerializableMap() map[interface{}]interface{} {
	copies := make(map[interface{}]interface{})
	m.rwmu.RLock()
	for k, v := range m.versionMap {
		copies[k] = v
	}
	m.rwmu.RUnlock()
	return copies
}
